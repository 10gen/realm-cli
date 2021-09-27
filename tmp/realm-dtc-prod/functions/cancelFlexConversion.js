exports = async (flexPlanId) => {
  const mapCustomerQueryProperties = (query) => context.functions.execute('mapCustomerQueryProperties', query);
  const client = context.services.get('mongodb-atlas');
  const customerCollection = client.db(context.values.get('DB_NAME')).collection('newcustomers');
  const flexPlansCollection = client.db(context.values.get('DB_NAME')).collection('flexplans');

  if (!flexPlanId) {
    throw new Error(`invalid cancelFlexConversion request: flexPlanId must be a valid id.  flexPlanId: ${flexPlanId}.`);
  }

  const flexPlan = await flexPlansCollection.findOne({ _id: BSON.ObjectId(flexPlanId) });
  if (!flexPlan) {
    throw new Error(`flexPlanId ${flexPlanId} not found in flexPlansCollection`);
  }

  const customerId = flexPlan.customer;
  if (!customerId) {
    throw new Error(`No customer id populated on flex plan ${flexPlan._id}`);
  }

  // cancel order
  const mostRecentInvoiceNumber = flexPlan.orders[flexPlan.orders.length - 1]?.invoiceNumber;
  if (mostRecentInvoiceNumber) {
    try {
      await context.functions.execute('cancelOrder', {
        invoiceNumber: mostRecentInvoiceNumber,
        cancelReason: 'Canceled Plan',
      });
    } catch (e) { // Order can't be canceled, just refund
      await context.functions.execute('refundOrder', {
        invoiceNumber: mostRecentInvoiceNumber,
        refundReason: 'Canceled Plan',
      });
    }
  }

  // remove plan from customer document
  let customer = await customerCollection.findOneAndUpdate(
    { _id: customerId },
    { $pull: { flexPlans: flexPlan._id } },
    { returnNewDocument: true },
  );

  // TODO: do we allow people with plans to start a new trial or does this always happen?
  let isCanceledConversion = false;
  if (customer.customerType === 'TrialToFlex' && customer.flexPlans.length === 0) {
    isCanceledConversion = true;
    customer = await customerCollection.findOneAndUpdate(
      { _id: customer._id },
      {
        $set: {
          noFollowup: new Date(),
        },
        $unset: {
          ConvertedFlex: 1,
          FlexFollowup: 1,
          StartFlex: 1,
        },
      },
      { returnNewDocument: true },
    );
  }

  // remove plan
  await flexPlansCollection.deleteOne({ _id: BSON.ObjectId(flexPlanId) });

  if (isCanceledConversion) {
    const eventPayload = {
      type: 'trial',
      action: 'conversion_canceled',
      customer: {
        id: customer._id.toString(),
      },
      timestamp: new Date(),
    };
    if (customer.query) {
      eventPayload.properties = mapCustomerQueryProperties(customer.query);
    }
    context.functions.execute('emitEvent', eventPayload);
  }
  context.functions.execute('emitKlaviyoTrack', flexPlan.email, 'CANCELED Flex Conversion');
  context.functions.execute('emitKlaviyoIdentify', flexPlan.email, { flexStatus: 'canceledConversion' });

  return customer;
};
