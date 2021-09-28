/*
  This function is run when a GraphQL Query is made requesting your
  custom field name. The return value of this function is used to
  populate the resolver generated from your Payload Type.

  This function expects the following input object:

  "input_type": {
        "properties": {
            "address": {
                "bsonType": "object",
                "properties": {
                    "address1": {
                        "bsonType": "string"
                    },
                    "address2": {
                        "bsonType": "string"
                    },
                    "city": {
                        "bsonType": "string"
                    },
                    "firstName": {
                        "bsonType": "string"
                    },
                    "lastName": {
                        "bsonType": "string"
                    },
                    "name": {
                        "bsonType": "string"
                    },
                    "state": {
                        "bsonType": "string"
                    },
                    "zip": {
                        "bsonType": "string"
                    }
                }
            },
            "customerId": {
                "bsonType": "string"
            },
            "discounts": {
                "bsonType": "array",
                "items": {
                    "additionalProperties": {
                        "amount": {
                            "bsonType": "int"
                        }
                    },
                    "properties": {
                        "code": {
                            "bsonType": "string"
                        },
                        "discountType": {
                            "bsonType": "string"
                        },
                        "value": {
                            "bsonType": "double"
                        }
                    },
                    "type": "object"
                }
            },
            "fulfillmentDelay": {
                "bsonType": "object",
                "properties": {
                    "duration": {
                        "bsonType": "int"
                    },
                    "durationUnit": {
                        "bsonType": "string",
                        "enum": [
                            "day",
                            "days"
                        ]
                    }
                }
            },
            "items": {
                "bsonType": "array",
                "items": {
                    "bsonType": "object",
                    "properties": {
                        "id": {
                            "bsonType": "string"
                        },
                        "quantity": {
                            "bsonType": "int"
                        }
                    }
                }
            },
            "orderType": {
                "bsonType": "string"
            },
            "source": {
                "bsonType": "string"
            },
            "shipping": {
                "bsonType": "int"
            }
        },
        "required": [
            "orderType",
            "customerId",
            "items"
        ],
        "title": "orderCreate",
        "type": "object"
    }

  returns: a created order
*/

exports = async (input) => {
  const MINIMUM_ORDER_TOTAL = context.values.get('MINIMUM_ORDER_TOTAL');
  const {
    customerId, source, orderContext, orderType,
  } = input;
  const client = context.services.get('mongodb-atlas');
  const customerColl = client.db('verbenergy').collection('newcustomers');
  const orderColl = client.db('verbenergy').collection('neworders');

  // Get Customer
  const customer = await customerColl.findOne({ _id: BSON.ObjectId(customerId) });
  if (!customer) {
    throw new Error(`Cannot process order for customer ${customerId}: Customer not found`);
  } else if (!customer.stripeID && !customer.customerId) {
    throw new Error(`Cannot process order for customer ${customerId}: Customer Stripe ID does not exist`);
  }

  // Create and charge order in a transaction
  const session = client.startSession();
  let order; let
    charge;
  try {
    await session.withTransaction(async () => {
      order = await context.functions.execute('createOrder', { ...input, customer, session });

      if (order.paid.creditUsed > 0) {
        await context.functions.execute('creditCustomer', { customerId: customer._id, credit: -order.paid.creditUsed });
      }

      if (order.paid.total >= MINIMUM_ORDER_TOTAL) {
        try {
          charge = await context.functions.execute('createOrderCharge', { order, customer });
        } catch (e) {
          // Failed Payment event triggers failed payment workflow
          const failedPaymentEvent = {
            type: 'customer',
            action: 'failed_payment',
            customer: {
              id: customer._id.toString(),
            },
            source,
            properties: {
              ...orderContext,
              wasRushed: customer.rushed,
              orderType,
            },
          };
          context.functions.execute('emitEvent', failedPaymentEvent);
          throw e;
        }
        try {
          await orderColl.updateOne(
            { _id: order._id },
            { $set: { stripeCharge: charge.id } },
            { session },
          );
          order.stripeCharge = charge.id;
        } catch (e) {
          console.error(`Error updating order ${order.invoiceNumber} with stripe charge ${charge.id}`, e);
        }
      }
    });
  } catch (e) {
    await session.abortTransaction();
    throw e;
  } finally {
    await session.endSession();
  }

  // update customer
  context.functions.execute('placedOrderCustomerUpdate', { fullDocument: order })
    .catch((e) => {
      console.error(`Error updating customer ${order.customer} for placed order ${order.invoiceNumber}`, e);
    });

  const event = context.functions.execute('composeOrderEvent', order, orderContext);
  context.functions.execute('emitEvent', event);

  // Klaviyo
  const klaviyoProperties = {
    ...event.properties,
    $value: event.properties.paid_total,
    $event_id: event.properties.invoiceNumber,
    crm: source === 'CRM',
  };
  context.functions.execute('emitKlaviyoTrack', customer.email, 'Placed Order', klaviyoProperties);

  // Google Analytics
  context.functions.execute('googleCollectOrder', { customer, order });

  // Slack Post
  const { SLACK_CHANNEL_ORDERS } = context.environment.values;
  if (SLACK_CHANNEL_ORDERS) {
    const orderMessage = {
      text: `Order ${order.invoiceNumber} - ${order.customerInfo.firstName} ${order.customerInfo.lastName}.
           ${order.orderType}. ${context.functions.execute('centsToDollars', order.paid.total)}`,
      channel: SLACK_CHANNEL_ORDERS,
    };
    context.functions.execute('slackPost', orderMessage)
      .catch((e) => {
        console.error(`Slack post failed for order ${order.invoiceNumber}`, e);
      });
  }

  // TODO: FB conversion

  return order;
};
