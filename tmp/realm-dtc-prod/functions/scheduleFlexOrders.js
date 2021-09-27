/*
  Function invoked by Flex Job Trigger. Finds flex plans with pending orders scheduled
  and creates flex orders.
*/

exports = async () => {
  const client = context.services.get('mongodb-atlas');
  const flexColl = client.db(context.values.get('DB_NAME')).collection('flexplans');

  const runDate = new Date();
  const plans = await flexColl.find({ status: 'active', nextOrder: { $lte: runDate } }, { _id: 1 }).limit(20).toArray();
  if (plans.length > 0) {
    // eslint-disable-next-line consistent-return
    const toProcess = (await Promise.all(plans.map(async ({ _id }) => {
      try {
        const result = await flexColl.updateOne(
          { _id, status: 'active', nextOrder: { $lte: runDate } },
          {
            $unset: {
              nextOrder: 1,
            },
          },
        );
        if (result.modifiedCount !== 1) {
          throw new Error(`Update for flex plan ${_id.toString()} did not result in a modification.`);
        }
        return _id.toString();
      } catch (e) {
        console.error(`Error scheduling flex order for ${_id.toString()}`, e);
      }
    }))).filter((entry) => entry);

    console.log(`Scheduling ${toProcess.length} flex orders`);
    toProcess.forEach((planId) => {
      context.functions.execute('processFlexOrder', { planId, source: 'CRM' })
        .catch((e) => {
          console.error(`Error processing flex order for plan ${planId}`, e);
        });
    });
  }
};
