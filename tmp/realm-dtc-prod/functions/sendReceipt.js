// eslint-disable-next-line consistent-return
exports = async (invoiceNumber) => {
  const _exec = execFnFor(context);
  const getOrder = _exec('getOrder');
  const composeOrderEvent = _exec('composeOrderEvent');
  const emitKlaviyoTrack = _exec('emitKlaviyoTrack');

  try {
    const order = await getOrder(invoiceNumber);
    const { properties } = await composeOrderEvent(order);
    await emitKlaviyoTrack(order.customerInfo.email, 'Send Receipt', properties);
    return order;
  } catch (error) {
    console.error(`sendReceipt for ${invoiceNumber} failed: '${error.message}'`, error);
  }
};
function execFnFor(context) {
  return (fnName) => (...args) => context.functions.execute(fnName, ...args);
}
