exports = async (input) => {
  const { flexPlanId, duration = 28 } = input;
  const moment = require('moment');

  // i think gql does this
  if (!flexPlanId) {
    throw new Error(`invalid skipFlexPlan request: flexPlanId must be a valid id.  flexPlanId: ${flexPlanId}.`);
  }
  const flexPlansCollection = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME')).collection('flexplans');

  const flexPlan = await flexPlansCollection.findOneAndUpdate(
    { _id: BSON.ObjectId(flexPlanId) },
    {
      $set: {
        nextText: moment().add(duration, 'days').hour(14).toDate(),
        status: 'active',
      },
      $unset: { nextOrder: 1 },
    },
    { returnNewDocument: true },
  );

  if (!flexPlan) {
    throw new Error(`flexPlan with id ${flexPlanId} not found)`);
  }

  context.functions.execute('emitEvent', {
    type: 'flex',
    action: 'skipped',
    customer: {
      id: flexPlan.customer.toString(),
    },
    properties: {
      typeId: flexPlan._id,
      daysSkipped: duration,
    },
  });

  return flexPlan;
};
