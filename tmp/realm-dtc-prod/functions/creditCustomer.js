/*
  MongoDB Realm function for crediting a customer

  input : {
    object: {
      customerId: ObjectID - the customer ID to credit
      credit: int - the credit value to apply
      source: string (optional) - event source for the credit
    }
  }

*/

exports = async ({ credit, customerId, source }) => {
  const client = context.services.get('mongodb-atlas');
  const customerColl = client.db('verbenergy').collection('newcustomers');
  const customer = await customerColl.findOne({ _id: customerId });
  if (!customer) {
    raiseCustomerNotFound(customerId);
  }
  const customerCredit = customer.credit ?? 0;
  const maxCredit = parseInt(Math.max(customerCredit + credit, 0) - customerCredit);

  const result = await customerColl.updateOne(
    { _id: customerId },
    { $inc: { credit: maxCredit } },
  );
  if (result.matchedCount !== 1) {
    raiseCustomerNotFound(customerId);
  } else if (result.modifiedCount !== 1) {
    console.warn(`Credit request for customer ${customerId} and value ${credit} did not result in an update`);

    return;
  }

  const event = {
    type: 'customer',
    action: credit > 0 ? 'credit' : 'debit',
    customer: {
      id: customerId.toString(),
    },
    properties: {
      credit: context.functions.execute('centsToDollars', maxCredit),
    },
  };
  if (source) {
    event.source = source;
  }
  context.functions.execute('emitEvent', event);

  function raiseCustomerNotFound(_id) {
    throw new Error(`Unable to credit customer: Customer ${_id} not found`);
  }
};
