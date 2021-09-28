/*
  This function composes an order event for a given order

*/

exports = (order, orderContext = {}) => {
  const centsToDollars = (value) => context.functions.execute('centsToDollars', value);

  let properties = {
    ...orderContext,
    typeId: order._id.toString(),
    orderType: order.orderType,
    invoiceNumber: order.invoiceNumber,
    customer_firstName: order.customerInfo.firstName,
    customer_lastName: order.customerInfo.lastName,
    customer_phone: order.customerInfo.customerPhone,
    shipping_firstName: order.customerInfo.address.firstName,
    shipping_lastName: order.customerInfo.address.lastName,
    shipping_address1: order.customerInfo.address.address1,
    shipping_address2: order.customerInfo.address.address2,
    shipping_city: order.customerInfo.address.city,
    shipping_state: order.customerInfo.address.state,
    shipping_zip: order.customerInfo.address.zip,
  };

  if (order.paid) {
    properties = {
      ...properties,
      paid_subtotal: centsToDollars(order.paid.subtotal),
      paid_shipping: centsToDollars(order.paid.shipping),
      paid_discountTotal: centsToDollars(order.paid.discountTotal),
      paid_creditUsed: centsToDollars(order.paid.creditUsed),
      paid_total: centsToDollars(order.paid.total),
      value: centsToDollars(order.paid.total),
    };
  }

  if (order.query) {
    Object.keys(order.query).forEach((key) => {
      properties[`query_${key}`] = order.query[key];
    });
  }

  if (order.items) {
    properties.itemSkus = order.items.map((item) => item.id);
    properties.items = order.items.map((item) => ({
      sku: item.id,
      name: item.name,
      price: centsToDollars(item.price),
      totalPrice: centsToDollars(item.totalPrice),
      quantity: item.quantity,
      metadata: {
        ...item.metadata,
        savings: centsToDollars(item.metadata?.savings),
      },
    }));

    const categories = new Set();
    const fulfillment = new Set();
    order.items.forEach((item) => {
      if (item.metadata?.categories?.length) {
        item.metadata.categories.forEach((category) => {
          categories.add(category);
        });
      }
      if (item.metadata?.fulfillment?.length) {
        item.metadata.fulfillment.forEach((fulfillmentItem) => {
          fulfillment.add(fulfillmentItem.sku);
        });
      } else fulfillment.add(item.sku);
    });
    properties.itemCategories = [...categories];
    properties.fulfillmentSkus = [...fulfillment];
  }

  if (order.discounts) {
    properties.discountCodes = order.discounts.map((discount) => discount.code);
    properties.discounts = order.discounts.map((discount) => ({
      code: discount.code,
      amount: centsToDollars(discount.amount),
      discountType: discount.discountType,
    }));
  }

  return {
    type: 'order',
    action: 'placed',
    customer: {
      id: order.customer.toString(),
    },
    timestamp: order.completionDate,
    source: order.source,
    properties,
  };
};
