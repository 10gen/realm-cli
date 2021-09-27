/*
  A Database Trigger will always call a function with a changeEvent.
  Documentation on ChangeEvents: https://docs.mongodb.com/manual/reference/change-events/

  Access the _id of the changed document:
  const docId = changeEvent.documentKey._id;

  Access the latest version of the changed document
  (with Full Document enabled for Insert, Update, and Replace operations):
  const fullDocument = changeEvent.fullDocument;

  const updateDescription = changeEvent.updateDescription;

  See which fields were changed (if any):
  if (updateDescription) {
    const updatedFields = updateDescription.updatedFields; // A document containing updated fields
  }

  See which fields were removed (if any):
  if (updateDescription) {
    const removedFields = updateDescription.removedFields; // An array of removed fields
  }

  Functions run by Triggers are run as System users
  and have full access to Services, Functions, and MongoDB Data.

  Access a mongodb service:
  const collection = context.services.get("mongodb-atlas")
  .db("verbenergy")
  .collection("newcustomers");
  const doc = collection.findOne({ name: "mongodb" });

  Note: In Atlas Triggers, the service name is defaulted to the cluster name.

  Call other named functions if they are defined in your application:
  const result = context.functions.execute("function_name", arg1, arg2);

  Access the default http client and execute a GET request:
  const response = context.http.get({ url: <URL> })

  Learn more about http client here: https://docs.mongodb.com/realm/functions/context/#context-http
*/
exports = async (changeEvent) => {
  const client = context.services.get('mongodb-atlas');
  const customerColl = client.db('verbenergy').collection('newcustomers');
  const { customer, invoiceNumber, completionDate } = changeEvent.fullDocument;
  const totalValue = changeEvent.fullDocument.paid.total;

  const update = {
    $set: {
      lastInteraction: new Date(),
    },
    $inc: {
      totalOrders: 1,
      totalValue,
    },
    $push: {
      orders: {
        invoiceNumber,
        completionDate,
      },
    },
  };

  const result = await customerColl.updateOne(
    { _id: customer, 'orders.invoiceNumber': { $ne: invoiceNumber } },
    update,
  );
  if (result.matchedCount !== 1) {
    console.warn(`Unable to update customer order data: Customer ${customer} not found or order ${invoiceNumber} already present in customer orders`);
  } else if (result.modifiedCount !== 1) {
    throw new Error(`Placed order customer ${customer} update for order ${invoiceNumber} did not result in an update`);
  } else {
    console.log(`Updated customer ${customer} for order ${invoiceNumber}`);
  }
};
