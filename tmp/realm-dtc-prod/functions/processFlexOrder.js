/*
  This function is run when a GraphQL Query is made requesting your
  custom field name. The return value of this function is used to
  populate the resolver generated from your Payload Type.
  This function expects the following input object:
  "input_type": {
        "properties": {
            "planId": {
                "bsonType": "string"
            },
            "source": {
                "bsonType": "string"
            }
        },
        "required": [
            "planId"
        ],
        "title": "processFlexOrder",
        "type": "object"
    }
  returns: a created flex order
*/

exports = async ({ planId, source }) => {
  const moment = require('moment');
  const client = context.services.get('mongodb-atlas');
  const dbName = context.values.get('DB_NAME');
  const flexColl = client.db(dbName).collection('flexplans');
  const customerColl = client.db(dbName).collection('newcustomers');

  const planObjectId = BSON.ObjectId(planId);
  const plan = await flexColl.findOne({ _id: planObjectId });
  if (plan) {
    if (plan.customer) {
      plan.discounts.forEach((discount) => {
        if (discount.discountType === 'mult' && discount.value > 1) {
          discount.value /= 100;
          if (discount.value > 1) {
            throw new Error(`Error processing flex order: mult discount ${discount.code} on plan ${planId} is malformed`);
          }
        }
      });
      try {
        const orderPayload = {
          customerId: plan.customer.toString(),
          items: plan.items,
          orderType: 'Flex',
          address: plan.shippingAddress,
          source,
          discounts: plan.discounts,
          shipping: plan.shippingPrice,
          orderContext: { orderPath: 'Flex Order', flexOrderCount: plan.orders.length + 1, flexTotalValue: plan.totalValue },
        };
        const order = await context.functions.execute('processOrder', orderPayload);

        const { flexDiscounts, totalPriceAdjustment } = plan.discounts.reduce(
          (result, discount) => {
            if (discount.flex) {
              result.flexDiscounts.push(discount);
            } else {
              result.totalPriceAdjustment += discount.amount;
            }
            return result;
          }, { flexDiscounts: [], totalPriceAdjustment: 0 },
        );
        const planUpdate = {
          $inc: {
            totalOrders: 1,
            totalValue: parseInt(order.paid.total),
            totalPrice: parseInt(totalPriceAdjustment),
          },
          $push: {
            orders: {
              invoiceNumber: order.invoiceNumber,
              completionDate: order.completionDate,
            },
          },
          $set: {
            rushed: false,
            discounts: flexDiscounts,
            nextText: moment().add(1, 'month').subtract(2, 'days').hour(14)
              .toDate(),
          },
          $unset: {
            nextOrder: 1,
          },
        };
        const updatedPlan = await flexColl.findOneAndUpdate(
          { _id: planObjectId },
          planUpdate,
          { returnNewDocument: true },
        );

        await customerColl.updateOne({ _id: plan.customer }, { $set: { failedFlex: 0 } });

        const klaviyoProperties = {
          orderNumber: updatedPlan.totalOrders,
          firstName: updatedPlan.firstName,
        };
        context.functions.execute('emitKlaviyoTrack', updatedPlan.email, 'Flex Order', klaviyoProperties);

        const appreciationEndDate = new Date(context.values.get('APPRECIATION_MONTH_END_DATETIME')).getTime();
        if (!Number.isNaN(appreciationEndDate) && appreciationEndDate > Date.now()) {
          context.functions.execute('emitKlaviyoTrack', updatedPlan.email, 'verbAppreciationReferral', klaviyoProperties);
        }

        // TODO: FB Conversion event

        return order;
      } catch (e) {
        console.error(`Error processing flex order for plan ${planId}`, e);
        if (e.code === 'FAILED_CHARGE') {
          const customer = await customerColl.findOneAndUpdate(
            { _id: plan.customer },
            { $inc: { failedFlex: 1 } },
            { returnNewDocument: true },
          );

          if (customer.failedFlex === 4) {
            await flexColl.updateOne({ _id: planObjectId }, { $set: { rushed: false } });
            await context.functions.execute('pauseFlexPlan', planId);
          } else {
            await flexColl.updateOne(
              { _id: planObjectId },
              {
                $set: {
                  rushed: false,
                  nextOrder: moment().add(3, 'days').toDate(),
                },
                $unset: {
                  nextText: 1,
                },
              },
            );
          }
        }
        throw e;
      }
    } else {
      throw new Error(`No customer set on plan ${planId}`);
    }
  } else {
    throw new Error(`Plan does not exist: ${planId}`);
  }
};
