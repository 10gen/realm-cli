/*
  Function invoked by Trial Conversion Trigger. Finds pending trial conversions and initiates them.
*/

exports = async () => {
  const client = context.services.get('mongodb-atlas');
  const customerColl = client.db(context.values.get('DB_NAME')).collection('newcustomers');

  const runDate = new Date();
  const customerQuery = { customerType: 'TrialToFlex', StartFlex: { $lt: runDate } };
  const customers = await customerColl.find(customerQuery, { _id: 1 }).limit(20).toArray();
  if (customers.length > 0) {
    // eslint-disable-next-line consistent-return
    const toProcess = (await Promise.all(customers.map(async ({ _id }) => {
      try {
        const result = await customerColl.updateOne(
          { _id, ...customerQuery },
          {
            $unset: {
              StartFlex: 1,
            },
          },
        );
        if (result.modifiedCount !== 1) {
          throw new Error(`Update for customer ${_id.toString()} did not result in a modification.`);
        }
        return _id.toString();
      } catch (e) {
        console.error(`Error scheduling trial conversion for customer ${_id.toString()}`, e);
      }
    }))).filter((entry) => entry);

    console.log(`Scheduling ${toProcess.length} trial conversions`);
    toProcess.forEach((customerId) => {
      context.functions.execute('convertTrial', { customerId, source: 'CRM' })
        .catch((e) => {
          console.error(`Error processing trial conversion for customer ${customerId}`, e);
        });
    });
  }
};
