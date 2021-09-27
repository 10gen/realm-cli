/*
  MongoDB Realm function for creating a Verb order

  Expects the same input provided to the processOrder function composed with an additional
  session parameter.

  returns: the created verb order
*/

exports = async (input) => {
  const moment = require('moment');
  const {
    customer, items, orderType, address, fulfillmentDelay, source, discounts, session, shipping,
  } = input;

  // TODO: env variables for collection names?
  const db = context.services.get('mongodb-atlas').db('verbenergy');
  const invoiceColl = db.collection('invoicenumbers');
  const orderColl = db.collection('neworders');

  // Build Cart
  const cart = await context.functions.execute('buildCart', {
    items, customer, orderType, source, discounts, shipping,
  });

  const shippingAddress = address || customer.shippingAddress;
  const shipDate = fulfillmentDelay
    ? moment().add(fulfillmentDelay.duration, fulfillmentDelay.durationUnit).toDate()
    : new Date();
  const invoiceNumber = await invoiceColl.findOneAndUpdate(
    {}, { $inc: { num: 1 } }, { returnNewDocument: true },
  );
  if (!invoiceNumber || !invoiceNumber.num) {
    throw new Error('Error generating invoice number.');
  }
  const order = {
    ...cart,
    customer: customer._id,
    customerInfo: {
      customerType: customer.customerType,
      firstName: customer.firstName,
      lastName: customer.lastName,
      customerPhone: customer.phone,
      email: customer.email,
      address: {
        ...shippingAddress,
        firstName: shippingAddress.firstName || customer.firstName,
        lastName: shippingAddress.lastName || customer.lastName,
      },
    },
    invoiceNumber: `VRB${invoiceNumber.num}`,
    completionDate: new Date(),
    shipping: {
      shippingType: 'ground',
      status: 'Pending',
      shipDate,
    },
  };

  // Save order to DB
  const orderResult = await orderColl.insertOne(order, { session });
  order._id = orderResult.insertedId;

  return order;
};
