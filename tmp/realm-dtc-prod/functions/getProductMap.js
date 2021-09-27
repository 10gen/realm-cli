/*
  MongoDB Realm helper function for generating a product map for a list of items

  input: {
    items - list of items
  }

  returns: a mapping of sku to product for each item in the input
*/

exports = async (items) => {
  const db = context.services.get('mongodb-atlas').db('verbenergy');
  const productColl = db.collection('products');
  const skus = items.map((item) => item.id);
  const products = await productColl.find({ sku: { $in: skus } }).toArray();
  if (products.length !== items.length) {
    throw new Error(`invalid order item sku in items: ${skus}`);
  }
  return products.reduce((acc, product) => {
    acc[product.sku] = product;
    return acc;
  }, {});
};
