exports = async (order, event = 'Flex Converted') => {
  try {
    const centsToDollars = (value) => context.functions.execute('centsToDollars', value);

    const email = order.customerInfo?.email;
    if (!email) {
      throw new Error('Email not populated on order.');
    }

    const klaviyoOrder = {
      invoiceNumber: order.invoiceNumber,
      customerInfo: order.customerInfo,
    };

    klaviyoOrder.items = order.items.map(({
      name, quantity, price, totalPrice,
    }) => ({
      name,
      quantity,
      price: centsToDollars(price),
      totalPrice: centsToDollars(totalPrice),
    }));

    klaviyoOrder.discounts = order.discounts.map(({ code, amount }) => ({
      code,
      amount: centsToDollars(amount),
    }));

    klaviyoOrder.paid = (({
      subtotal, shipping, discountTotal, total,
    }) => ({
      subtotal: centsToDollars(subtotal),
      shipping: centsToDollars(shipping),
      discountTotal: centsToDollars(discountTotal),
      total: centsToDollars(total),
    }))(order.paid);

    context.functions.execute('emitKlaviyoTrack', email, event, klaviyoOrder);
  } catch (e) {
    console.error(`Klaviyo Start Flex failed for ${order._id}`, e);
  }
};
