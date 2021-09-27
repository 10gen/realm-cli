exports = async (customerId) => {
  if (!customerId) {
    throw new Error(`invalid cancelTrial request: customerId must be a valid id.  customerId: ${customerId}.`);
  }
  const customersCollection = context.services.get('mongodb-atlas').db(context.values.get('DB_NAME')).collection('newcustomers');
  const mapCustomerQueryProperties = (query) => context.functions.execute('mapCustomerQueryProperties', query);
  const currDate = new Date();
  const updated = await customersCollection.findOneAndUpdate(
    { _id: BSON.ObjectId(customerId) },
    {
      $set: {
        NoFollowup: currDate,
      },
      $unset: {
        StartFlex: 1,
        FlexFollowup: 1,
      },
    },
    { returnNewDocument: true },
  );

  if (updated) {
    const eventPayload = {
      type: 'trial',
      action: 'canceled',
      customer: {
        id: customerId.toString(),
      },
      timestamp: currDate,
      properties: {
        typeId: customerId,
      },
    };

    if (updated.query) {
      eventPayload.properties = {
        ...eventPayload.properties,
        ...mapCustomerQueryProperties(updated.query),
      };
    }
    context.functions.execute('emitEvent', eventPayload);

    context.functions.execute('emitKlaviyoTrack', updated.email, 'Trial Canceled');
    context.functions.execute('emitKlaviyoIdentify', updated.email);
  }
  return updated;
};
