exports = async (invoiceNumber) => {
  const DB_NAME = context.values.get('DB_NAME');
  const ordersCollection = context.services.get('mongodb-atlas').db(DB_NAME).collection('neworders');
  const order = await ordersCollection.findOne({ invoiceNumber });

  if (!order) {
    throw new Error(`Order ${invoiceNumber} not found`);
  }
  return order;
};
