/*
  MongoDB Realm function for converting a trial

  input: {
    customerId - id of converting customer
    source - order source
  }

*/

exports = async ({ customerId, source }) => {
  const moment = require('moment');
  const client = context.services.get('mongodb-atlas');
  const dbName = context.values.get('DB_NAME');
  const flexColl = client.db(dbName).collection('flexplans');
  const customerColl = client.db(dbName).collection('newcustomers');

  const customer = await customerColl.findOne({ _id: BSON.ObjectId(customerId) });
  if (!customer) {
    throw new Error(`Customer ${customerId} not found`);
  } else if (!customer.stripeID && !customer.customerId) {
    throw new Error(`Customer ${customerId} Stripe ID is undefined`);
  }

  const items = [];
  if (customer.flexDefault?.length > 0) {
    // TODO: does this even exist anymore?
    if (customer.flexDefault[0].name === 'Variety Sampler') {
      items.push(
        {
          id: 'pb-sampler',
          quantity: 1,
        },
        {
          id: 'mb-sampler',
          quantity: 1,
        },
        {
          id: 'sc-sampler',
          quantity: 1,
        },
        {
          id: 'cc-sampler',
          quantity: 1,
        },
      );
    } else {
      customer.flexDefault.forEach(({ id, quantity }) => {
        if (id && quantity) {
          items.push({ id, quantity });
        } else {
          throw new Error(`Malformed flex default item encountered for customer ${customerId}`);
        }
      });
    }
  } else {
    throw new Error(`Flex default not configured for customer ${customerId}`);
  }

  const orderPayload = {
    customerId,
    items,
    orderType: 'Flex',
    address: customer.shippingAddress,
    source,
    orderContext: {
      orderPath: 'Trial Conversion',
    },
  };
  if (!customer.rushed) {
    orderPayload.fulfillmentDelay = { duration: 1, durationUnit: 'days' };
  }

  let order;
  try {
    order = await context.functions.execute('processOrder', orderPayload);
  } catch (e) {
    if (e.code === 'FAILED_CHARGE') {
      const customerUpdate = {
        $unset: {
          rushed: 1,
        },
        $inc: {
          failedStart: 1,
        },
      };
      if (customer.failedStart + 1 >= 2) {
        customerUpdate.$set = {
          NoFollowup: new Date(),
        };
        customerUpdate.$unset = {
          ...customerUpdate.$unset,
          StartFlex: 1,
          FlexFollowup: 1,
        };
      } else {
        customerUpdate.$set = {
          StartFlex: moment().add(2, 'days').hour(14).toDate(),
        };
      }
      await customerColl.updateOne({ _id: customer._id }, customerUpdate);
    } else {
      // Order Processing Error
      const customerUpdate = {
        $unset: {
          rushed: 1,
        },
        $set: {
          StartFlex: moment().add(2, 'days').hour(14).toDate(),
        },
      };
      await customerColl.updateOne({ _id: customer._id }, customerUpdate);
    }

    throw e;
  }

  if (!order) {
    throw new Error(`Fatal error with trial conversion for customer ${customerId}: there is likely a bug in the code.`);
  }

  const newPlan = {
    firstName: customer.firstName,
    lastName: customer.lastName,
    phone: customer.phone,
    customer: customer._id,
    email: customer.email,
    items: order.items,
    discounts: [],
    shippingAddress: customer.shippingAddress,
    totalPrice: parseInt(order.paid.subtotal + order.paid.shipping),
    nextText: moment().add(1, 'month').subtract(2, 'days').hour(14)
      .toDate(),
    started: order.completionDate,
    status: 'active',
    source,
    totalValue: order.paid.total,
    totalOrders: 1,
    orders: [{ invoiceNumber: order.invoiceNumber, completionDate: order.completionDate }],
  };
  order.discounts?.forEach((discount) => {
    if (discount.flex) {
      newPlan.discounts.push(discount);
      newPlan.totalPrice -= parseInt(discount.amount);
    }
  });

  try {
    const planResult = await flexColl.insertOne(newPlan);
    if (!planResult.insertedId) {
      throw new Error('Flex Plan insert failed: WriteResult did not include insertedId');
    }
    newPlan._id = planResult.insertedId;
  } catch (e) {
    const message = `Error creating flex plan for customer ${customer._id}: ${e.message}`;
    console.error(message, e);
    throw new Error(message);
  }

  if (!newPlan._id) {
    throw new Error(`Error converting trial for customer ${customer._id} with order ${order._id}: 
      Plan creation executed successfully but plan _id is undefined. There is likely a bug.`);
  }

  try {
    await customerColl.updateOne(
      { _id: customer._id },
      {
        $push: {
          flexPlans: newPlan._id,
        },
        $unset: {
          rushed: 1,
          FlexFollowup: 1,
          NoFollowup: 1,
          StartFlex: 1,
        },
        $set: {
          ConvertedFlex: new Date(),
        },
      },
    );
  } catch (e) {
    const message = `Error updating customer flex plans with plan ${newPlan._id}: ${e.message}`;
    console.error(message, e);
    throw new Error(message);
  }

  let mappedQueryProperties;
  if (customer.query) {
    mappedQueryProperties = context.functions.execute('mapCustomerQueryProperties', customer.query);
  }
  const trialConvertedEvent = {
    type: 'trial',
    action: 'converted',
    customer: {
      id: customer._id.toString(),
    },
    timestamp: newPlan.started,
  };
  if (mappedQueryProperties) {
    trialConvertedEvent.properties = mappedQueryProperties;
  }
  context.functions.execute('emitEvent', trialConvertedEvent);

  const flexCreatedEvent = {
    type: 'flex',
    action: 'created',
    customer: {
      id: customer._id.toString(),
    },
    timestamp: newPlan.started,
    properties: {
      typeId: newPlan._id,
    },
  };
  if (mappedQueryProperties) {
    flexCreatedEvent.properties = { ...trialConvertedEvent.properties, mappedQueryProperties };
  }
  context.functions.execute('emitEvent', flexCreatedEvent);

  context.functions.execute('emitKlaviyoIdentify', customer.email, { flexStatus: newPlan.status });
  context.functions.execute('emitKlaviyoStartFlex', order);

  // TODO: FB Conversion Event
};
