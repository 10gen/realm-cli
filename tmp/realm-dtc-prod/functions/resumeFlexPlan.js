exports = async (flexPlanId) => {
  const moment = require('moment');

  if (!flexPlanId) {
    throw new Error(`invalid resumeFlexResume request: flexPlanId must be a valid id.  flexPlanId: ${flexPlanId}.`);
  }
  const flexPlansCollection = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME')).collection('flexplans');

  const flexPlan = await flexPlansCollection.findOneAndUpdate(
    { _id: BSON.ObjectId(flexPlanId), status: { $ne: 'active' } },
    {
      $set: {
        status: 'active',
        resumedOn: new Date(),
        nextText: moment().add(28, 'days').hour(14).toDate(),
      },
      $unset: {
        nextOrder: 1,
      },
    },
    { returnNewDocument: true },
  );

  if (!flexPlan) {
    throw new Error(`flexPlan with id ${flexPlanId} not found or in invalid state (plan.status must not already be active)`);
  }

  const customersCollection = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME')).collection('newcustomers');
  let customer = await customersCollection.findOne(
    { _id: flexPlan.customer },
  );

  if (!customer) {
    throw new Error(`flexPlan customer missing or not present: flexPlan ID - ${flexPlan._id} plan.customer.id - ${flexPlan.customer}`);
  }

  if (customer.customerType === 'TrialToFlex' && !customer.ConvertedFlex) {
    customer = await customersCollection.findOneAndUpdate(
      { _id: customer._id },
      {
        $unset: { NoFollowup: 1, FlexFollowup: 1, StartFlex: 1 },
        $set: { ConvertedFlex: flexPlan.started },
      },
    );
  }

  context.functions.execute('emitEvent', {
    type: 'flex',
    action: 'resumed',
    customer: {
      id: customer._id.toString(),
    },
    timestamp: flexPlan.resumedOn,
    properties: {
      typeId: flexPlan._id,
    },
  });

  context.functions.execute('emitKlaviyoTrack', flexPlan.email, 'Flex Resumed');
  context.functions.execute('emitKlaviyoIdentify', flexPlan.email, { flexStatus: flexPlan.status });

  return flexPlan;
};
