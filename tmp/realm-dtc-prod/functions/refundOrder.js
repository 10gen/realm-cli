exports = async ({ invoiceNumber, refundReason, refundAmount = 0 }) => {
  const querystring = require('querystring');

  const ordersCollection = context.services.get('mongodb-atlas').db(context.values.get('DB_NAME')).collection('neworders');
  const order = await ordersCollection.findOne({ invoiceNumber });
  if (!order) {
    throw new Error(`Order ${invoiceNumber} not found`);
  }

  const priorRefundTotal = order.refund?.filter(({ status }) => status === 'succeeded')
    .reduce((total, refund) => total + refund.amount, 0) ?? 0;
  const refundableTotal = order.paid.total - priorRefundTotal;
  const toRefund = refundAmount > 0 ? refundAmount : refundableTotal;
  if (toRefund > refundableTotal) {
    throw new Error(`Requested refund of ${toRefund} exceeds the value of this order: ${invoiceNumber}`);
  }

  const headers = {
    Authorization: [`Bearer ${context.values.get('STRIPE_SECRET_KEY')}`],
    'Content-Type': ['application/x-www-form-urlencoded'],
  };

  const body = {
    charge: order.stripeCharge,
    amount: toRefund,
  };

  const response = await context.http.post({
    url: 'https://api.stripe.com/v1/refunds',
    headers,
    body: querystring.stringify(body),
  });
  if (response.statusCode !== 200) {
    throw new Error(`stripe refund request ${JSON.stringify(body)} failed with error: ${JSON.stringify(response.body.text())}`);
  }

  const refund = EJSON.parse(response.body.text());
  const update = {
    $push: {
      refund: {
        amount: refund.amount,
        reason: refundReason,
        stripeRefund: refund.id,
        status: refund.status,
        refundDate: new Date(),
      },
    },
  };
  await ordersCollection.updateOne(
    { _id: order._id },
    update,
  );

  const customerCollection = context.services.get('mongodb-atlas').db(context.values.get('DB_NAME')).collection('newcustomers');
  // todo: richer partial refunds,
  // e.g. decrement totalOrders if partial refunds add up to totalOrder
  // for now if refundAmount is empty and refunding whole order decrement customer.totalOrders
  const customerUpdate = refundAmount ? { $inc: { totalValue: -refund.amount } }
    : { $inc: { totalValue: -refund.amount, totalOrders: -1 } };

  await customerCollection.updateOne(
    { _id: order.customer },
    customerUpdate,
  );

  return true;
};
