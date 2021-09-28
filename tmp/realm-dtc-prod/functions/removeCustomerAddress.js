/*
 Remove Customer Address Book Records.
*/
exports = async ({ addressBookIds, customerId }) => {
  const collection = context.services.get('mongodb-atlas').db('verbenergy').collection('newcustomers');
  const _id = new BSON.ObjectId(customerId);
  const bsonIds = addressBookIds.map((strId) => new BSON.ObjectId(strId));
  const query = { _id };
  const update = {
    $pull: { addressBook: { _id: { $in: bsonIds } } },
  };
  const options = { returnNewDocument: true };
  const customer = await collection.findOneAndUpdate(query, update, options);
  const serializedCustomer = JSON.stringify(customer); // https://docs.mongodb.com/realm/functions/#behavior
  console.log('UPDATED CUSTOMER: ', serializedCustomer);
  return customer;
};
