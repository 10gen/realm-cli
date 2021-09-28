/*
  MongoDB Realm function for manually creating a new flex plan

  input: {
    items - list of order items
    discounts - list of input discounts
    customerId - id of converting customer
    source - order source
  }

  returns: an updated customer after placing the order and creating the plan
*/

exports = async (input) => {
  const {
    customerId, source, discounts = [], items,
  } = input;
  const db = context.services.get('mongodb-atlas').db('verbenergy');
  const flexColl = db.collection('flexplans');
  const moment = require('moment');
  const customerColl = db.collection('newcustomers');
  let customer = await customerColl.findOne({ _id: BSON.ObjectId(customerId) });

  if (!customer) {
    throw new Error(`Cannot process order for customer ${customerId}: Customer not found`);
  } else if (!customer.stripeID && !customer.customerId) {
    throw new Error(`Cannot process order for customer ${customerId}: Customer Stripe ID does not exist`);
  }
  const orderPayload = {
    customerId,
    items,
    orderType: 'Flex',
    address: customer.shippingAddress,
    source,
    discounts,
    orderContext: {
      orderPath: 'Flex Order',
      flexOrderCount: 1,
    },
  };
  if (!customer.rushed) {
    orderPayload.fulfillmentDelay = { duration: 1, durationUnit: 'days' };
  }
  const order = await context.functions.execute('processOrder', orderPayload);
  const currDate = moment().toDate();
  const nextDate = moment(currDate).add(1, 'month').toDate();
  const { shippingAddress } = customer;
  let totalPrice = parseInt(order.paid.shipping);
  order.items.forEach((item) => {
    totalPrice += parseInt(item.totalPrice);
  });
  const newPlan = {
    firstName: customer.firstName,
    lastName: customer.lastName,
    phone: customer.phone,
    customer: customer._id,
    email: customer.email,
    items: order.items,
    discounts: order.discounts,
    shippingAddress,
    totalPrice,
    nextText: nextDate,
    started: currDate,
    status: 'active',
    source,
    totalValue: order.paid.total,
    totalOrders: 1,
    orders: [{
      invoiceNumber: order.invoiceNumber,
      completionDate: order.completionDate,
    }],
    rushed: false,
  };

  const newPlanResult = await flexColl.insertOne(newPlan);
  newPlan._id = newPlanResult.insertedId;
  customer = await customerColl.findOneAndUpdate(
    { _id: BSON.ObjectId(customerId) },
    {
      $push: {
        flexPlans: newPlan._id,
      },
      $set: {
        ConvertedFlex: currDate,
      },
      $unset: {
        FlexFollowup: 1,
        StartFlex: 1,
      },
    },
    { returnNewDocument: true },
  );
  context.functions.execute('emitKlaviyoIdentify', newPlan.email, { flexStatus: newPlan.status });
  context.functions.execute('emitKlaviyoStartFlex', order);
  const eventPayload = {
    type: 'flex',
    action: 'created',
    customer: {
      id: customerId,
    },
    timestamp: currDate,
    properties: {
      typeId: newPlan._id,
    },
  };
  context.functions.execute('emitEvent', eventPayload);

  return customer;
};
