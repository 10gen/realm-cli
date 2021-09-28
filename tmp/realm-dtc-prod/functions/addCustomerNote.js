exports = async ({ body, customerId, status = 'Display' }) => {
  /*
    Accessing application's values:
    var x = context.values.get("value_name");

    Accessing a mongodb service:
    var collection = context.services.get("mongodb-atlas").db("dbname").collection("coll_name");
    var doc = collection.findOne({owner_id: context.user.id});

    To call other named functions:
    var result = context.functions.execute("function_name", arg1, arg2);

    Try running in the console below.
  */
  const collection = context.services.get('mongodb-atlas').db('verbenergy').collection('newcustomers');
  const _id = new BSON.ObjectId(customerId);
  const customer = await collection.findOneAndUpdate({ _id }, {
    $push: { notes: { body, status, _id: new BSON.ObjectId() } },
  }, { returnNewDocument: true });
  // const customer = await collection.findOne({ _id });
  console.log('UPDATED CUSTOMER', customer);
  return JSON.parse(JSON.stringify(customer));
};
