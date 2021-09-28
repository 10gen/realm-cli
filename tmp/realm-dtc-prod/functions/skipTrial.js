exports = async (input) => {
  const { customerId, duration = 28 } = input;
  const moment = require('moment');
  const skipDate = moment().add(duration, 'days').hour(14).toDate();
  const mapCustomerQueryProperties = (query) => context.functions.execute('mapCustomerQueryProperties', query);
  if (!customerId) {
    throw new Error(`invalid skipTrial request: customerId must be a valid id.  customerId: ${customerId}.`);
  }
  const customerCollection = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME')).collection('newcustomers');

  const updated = await customerCollection.findOneAndUpdate(
    { _id: BSON.ObjectId(customerId) },
    {
      $set: {
        FlexFollowup: skipDate,
      },
      $unset: {
        StartFlex: 1,
        NoFollowup: 1,
      },
    },
    { returnNewDocument: true },
  );

  if (!updated) {
    throw new Error(`customer with id ${customerId} not found`);
  }
  const eventPayload = {
    type: 'trial',
    action: 'skipped',
    customer: {
      id: customerId,
    },
    properties: {
      daysSkipped: duration,
    },
  };

  if (updated.query) {
    eventPayload.properties = {
      ...eventPayload.properties,
      ...mapCustomerQueryProperties(updated.query),
    };
  }
  context.functions.execute('emitEvent', eventPayload);

  return updated;
};
