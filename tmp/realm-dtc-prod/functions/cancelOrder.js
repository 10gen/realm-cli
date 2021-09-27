exports = async ({ invoiceNumber, cancelReason = 'Canceled Order' }) => {
  const orderColl = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME'))
    .collection('neworders');

  const { matchedCount, modifiedCount } = await orderColl.updateOne(
    { invoiceNumber, 'shipping.status': { $in: ['Pending', 'On Hold'] } },
    { $set: { 'shipping.status': 'Canceled' } },
  );
  if (matchedCount !== 1) {
    throw new Error(`Order ${invoiceNumber} not found or has invalid order status for cancel.`);
  }
  if (modifiedCount !== 1) {
    throw new Error(`Error modifying order ${invoiceNumber}.`);
  }

  try {
    await context.functions.execute('refundOrder', { invoiceNumber, refundReason: cancelReason });
  } catch (e) {
    throw new Error(`successfully canceled ${invoiceNumber} but failed to refund: ${e.message}`);
  }

  return true;
};
