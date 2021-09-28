/*
  This function constructs a cart for a given order request.

  TODO: Make a custom resolver for this.

  returns: a subset of a newOrder object representing the order cart
*/

exports = async (input) => {
  const MINIMUM_ORDER_TOTAL = context.values.get('MINIMUM_ORDER_TOTAL');
  const DEFAULT_SHIPPING_PRICE = 495;
  const NO_PASSTHROUGH_SHIPPING = 'NO_PASSTHROUGH_SHIPPING';
  const db = context.services.get('mongodb-atlas').db('verbenergy');
  const customerColl = db.collection('newcustomers');
  const {
    items: itemsInput,
    customerId,
    orderType,
    source,
    discounts: discountsInput = [],
    shipping: shippingInput = NO_PASSTHROUGH_SHIPPING,
  } = input;
  let { customer } = input;

  // fetch if no customer object
  if (!customer) {
    if (!customerId) raiseNeedCustomerId();

    customer = await customerColl.findOne({ _id: BSON.ObjectId(customerId) });

    if (!customer) raiseCustomerNotFound(customerId);
  }

  const productMap = await context.functions.execute('getProductMap', itemsInput);
  let hasFreeShippingProduct = false;

  const items = itemsInput.map((item) => {
    const product = productMap[item.id];
    if (!product) raiseProductNotFound(item);

    const productPrice = parseInt(product.price.value);
    if (isNaN(productPrice)) raiseProductPriceInvalid(item);

    if (isNaN(item.quantity)) raiseItemQuantityInvalid(item);

    const itemTotal = productPrice * item.quantity;

    // check if item triggers free shipping and is present (quantity != 0)
    hasFreeShippingProduct |= product.freeShipping && item.quantity;

    let savings = 0;
    if (product.price.strikethrough || (product.price.value < product.price.original)) {
      const highestPrice = Math.max(product.price.strikethrough, product.price.original);
      savings = highestPrice - product.price.value;
    }

    const orderItem = {
      id: product.sku,
      product: product._id,
      name: product.name,
      quantity: item.quantity,
      price: productPrice,
      totalPrice: itemTotal,
      metadata: {
        categories: product.categories?.map((category) => category?.id),
        fulfillment: product.fulfillment,
        thumbnail: product.cms?.media?.cartThumbnail,
        savings: parseInt(savings),
      },
    };

    return orderItem;
  });

  let shipping;
  if (shippingInput === NO_PASSTHROUGH_SHIPPING) { // (if input.shipping is undefined)
    if (orderType === 'Flex' || hasFreeShippingProduct) {
      shipping = 0;
    } else {
      shipping = DEFAULT_SHIPPING_PRICE;
    }
  } else {
    shipping = parseInt(shippingInput);
  }

  // add to subtotal

  const subtotal = items.reduce((acc, { totalPrice }) => acc + parseInt(totalPrice), 0);

  const fullSubtotal = subtotal + shipping;

  let total = 0;

  // Get Discounts
  const discounts = await context.functions.execute('getCartDiscounts', {
    items,
    discounts: discountsInput,
    productMap,
    orderType,
    subtotal,
  });

  let discountTotal = discounts.reduce((acc, { amount }) => acc + parseInt(amount), 0);

  // check if discounts exceed total
  if (fullSubtotal >= discountTotal) {
    total = fullSubtotal - discountTotal;
  } else {
    discountTotal = fullSubtotal;
  }

  // Check if customer has available credit
  let creditUsed = 0;
  const customerCredit = parseInt(customer.credit) || 0;
  if (customerCredit > 0) {
    creditUsed = Math.min(total, customerCredit);
    total -= creditUsed;
  }

  if (total < MINIMUM_ORDER_TOTAL) {
    total = 0;
  }

  const cart = {
    items,
    orderType,
    source,
    discounts,
    paid: {
      subtotal: parseInt(subtotal),
      shipping: parseInt(shipping),
      discountTotal: parseInt(discountTotal),
      creditUsed: parseInt(creditUsed),
      total: parseInt(total),
    },
  };

  return cart;
};
function raiseItemQuantityInvalid(item) {
  throw new Error(`Invalid quantity '${item.quantity}' given for product '${item.id}'`);
}

function raiseProductPriceInvalid(item) {
  throw new Error(`Product pricing for '${item.id}' not found`);
}

function raiseProductNotFound(item) {
  throw new Error(`Product with sku '${item.id}' not found`);
}

function raiseCustomerNotFound(customerId) {
  throw new Error(`Invalid cart request: customer '${customerId}' not found`);
}

function raiseNeedCustomerId() {
  throw new Error('Invalid cart request: one of customer or customerId must be provided');
}
