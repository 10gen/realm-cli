/*
  MongoDB Realm function for getting discounts to apply to a list of items

  input: {
    items - list of order items
    discounts - list of input discounts
    subtotal (optional) - cart subtotal
  }

  returns: a list of discounts associated with the input order item list
*/

exports = async (input) => {
  const AUTO_DISCOUNTS = new Set(context.values.get('AUTO_DISCOUNTS') || ['HOLIDAYBUNDLEPROMO']);
  const NON_DISCOUNTABLE_SKUS = new Set(context.values.get('NON_DISCOUNTABLE_SKUS'));
  const { items, discounts = [] } = input;
  const productMap = input.productMap || await context.functions.execute('getProductMap', items);

  let discountableTotal = 0;
  let hasSamplerPouch = false;
  let hasStarterKit = false;
  const appreciationProductCount = 0;
  const appreciationTotalPrice = 0;
  let pslPrice = 0;
  let pslQuantity = 0;
  let hasPSL = false;

  items.forEach((item) => {
    const product = productMap[item.id];
    if (!product) {
      throw new Error(`Error computing cart discounts: Product mapping not found for ${item.id}`);
    }

    if (item.id === 'ps-pouch') {
      pslPrice = product.price.value;
      hasPSL = true;
      pslQuantity = item.quantity;
    }

    // check for product presence in cart
    hasSamplerPouch |= item.id === 'sampler-pouch';
    hasStarterKit |= item.id === 'starter-kit';
    const hasOverrideDiscount = false;

    if (!(NON_DISCOUNTABLE_SKUS.has(item.id)
    || product.price.value < product.price.original
    || hasOverrideDiscount)) {
      discountableTotal += item.totalPrice;
    }
  });

  const cartDiscounts = [];
  const addToCartDiscounts = (discount) => (rawAmount) => {
    cartDiscounts.push({ ...discount, amount: parseInt(rawAmount) });
  };

  discounts.forEach((discount) => {
    if (AUTO_DISCOUNTS.has(discount.code)) return;

    const pushToCart = addToCartDiscounts(discount);
    let markdownRate = 1;
    switch (discount.discountType) {
      case 'Trial':
      case 'Referral':
        if (hasSamplerPouch) {
          pushToCart(discount.value * 100);
        }
        break;
      case 'mult':
        markdownRate = discount.value;
        if (markdownRate > 1) {
          markdownRate /= 100;
          if (markdownRate > 1) {
            throw new Error('Error computing cart discounts: mult discount value is invalid.');
          }
        }
        pushToCart(discountableTotal * markdownRate);
        break;
      case 'minus':
        pushToCart(discount.value * 100);
        break;
      case 'starter':
        if (hasStarterKit) {
          pushToCart(discount.value * 100);
        }
        break;
      case 'ogverbfam':
        if (discountableTotal >= 5000) {
          pushToCart(discountableTotal * 0.15);
        } else if (discountableTotal >= 3000) {
          pushToCart(discountableTotal * 0.10);
        }
        break;
      default:
        console.warn(`unknown discount type ${discount.discountType} provided`);
    }
  });

  if (appreciationProductCount > 0) {
    const value = 0.15;
    addToCartDiscounts({
      code: 'VERBAPPRECIATION',
      discountType: 'mult',
      value,
    })(value * appreciationTotalPrice);
  }
  if (hasPSL) {
    const value = 0.2;
    addToCartDiscounts({
      code: 'PSLFALL2021',
      discountType: 'mult',
      value,
    })(value * pslPrice * pslQuantity);
  }

  return cartDiscounts;
};
