exports = async (flexPlanId) => {
  if (!flexPlanId) {
    throw new Error(`invalied pauseFlexPlan request: flexPlanId must be a valid id.  flexPlanId: ${flexPlanId}.`);
  }

  const flexPlansCollection = context.services.get('mongodb-atlas').db(context.values.get('DB_NAME')).collection('flexplans');
  const flexPlan = await flexPlansCollection.findOneAndUpdate(
    { _id: BSON.ObjectId(flexPlanId) },
    { $set: { status: 'paused', pausedOn: new Date(), rushed: false }, $unset: { nextText: '', nextOrder: '' } },
    { returnNewDocument: true },
  );

  const customerId = flexPlan.customer;
  if (customerId) {
    context.functions.execute('emitEvent', {
      type: 'flex',
      action: 'paused',
      customer: {
        id: customerId.toString(),
      },
      timestamp: flexPlan.pausedOn,
      properties: {
        typeId: flexPlan._id,
      },
    });

    context.functions.execute('emitKlaviyoTrack', flexPlan.email, 'Flex Paused');
    context.functions.execute('emitKlaviyoIdentify', flexPlan.email, { flexStatus: flexPlan.status });
  }
  return flexPlan;
};
