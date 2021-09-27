/*
  MongoDB Realm function for merge account
  information over to new customer
  input: {
    toArchiveCustomerId - new "shell" customer customerId
    toKeepCustomerId - old current customer customerId
  }
  returns: an updated customer after merging the shell account
*/
exports = async (input) => {
  const { toKeepCustomerId, toArchiveCustomerId } = input;
  const db = context.services.get('mongodb-atlas').db('verbenergy');
  const customerColl = db.collection('newcustomers');
  const messageColl = db.collection('messages');
  const flexPlanColl = db.collection('flexplans');
  const ticketColl = db.collection('tickets');
  let updatedCustomer = null;

  if (!toKeepCustomerId || !toArchiveCustomerId) {
    console.error(`Error with customerIds: ${toArchiveCustomerId} and  ${toKeepCustomerId}`);
    throw new Error(`Error with customerIds: ${toArchiveCustomerId} and  ${toKeepCustomerId}`);
  }
  try {
    // pull the old (toKeep) and new (toArchive) customer accounts
    const toKeepCustomer = await customerColl.findOne({ _id: BSON.ObjectId(toKeepCustomerId) });
    const toArchiveCustomer = await customerColl.findOne(
      { _id: BSON.ObjectId(toArchiveCustomerId) },
    );

    if (!(toKeepCustomer && toArchiveCustomer && toKeepCustomer.phone && toArchiveCustomer.phone)) {
      throw new Error(`Error merging customers: ${toArchiveCustomerId} and  ${toKeepCustomerId}`);
    }
    // update selected fields to prep for merge
    if (toKeepCustomer.lastInteraction && toArchiveCustomer.lastInteraction) {
      toKeepCustomer.lastInteraction = toKeepCustomer.lastInteraction
        > toArchiveCustomer.lastInteraction
        ? toKeepCustomer.lastInteraction : toArchiveCustomer.lastInteraction;
    } else {
      toKeepCustomer.lastInteraction = toKeepCustomer.lastInteraction
      || toArchiveCustomer.lastInteraction || new Date();
    }
    if (toKeepCustomer.created && toArchiveCustomer.created) {
      toKeepCustomer.created = toKeepCustomer.created > toArchiveCustomer.created
        ? toKeepCustomer.created : toArchiveCustomer.created;
    } else {
      toKeepCustomer.created = toKeepCustomer.created || toArchiveCustomer.created || new Date();
    }
    if ((toKeepCustomer.lastText && toKeepCustomer.lastText.timestamp)
        && (toArchiveCustomer.lastText && toArchiveCustomer.lastText.timestamp)) {
      toKeepCustomer.lastText = toKeepCustomer.lastText.timestamp
        < toArchiveCustomer.lastText.timestamp
        ? toArchiveCustomer.lastText : toKeepCustomer.lastText;
    } else if ((toKeepCustomer.lastText && toKeepCustomer.lastText.timestamp)
        || (toArchiveCustomer.lastText && toArchiveCustomer.lastText.timestamp)) {
      toKeepCustomer.lastText = toKeepCustomer.lastText
        ? toKeepCustomer.lastText : toArchiveCustomer.lastText;
    }
    // update all old messages with new phone number
    await messageColl.updateMany(
      { customerId: BSON.ObjectId(toKeepCustomerId) },
      {
        $set: {
          customer: toArchiveCustomer.phone,
        },
      },
    );
    // update all new messages with old customer id
    await messageColl.updateMany(
      { customerId: BSON.ObjectId(toArchiveCustomerId) },
      {
        $set: {
          customerId: BSON.ObjectId(toKeepCustomerId),
        },
      },
    );
    // update flex plan(s) with new phone number
    await flexPlanColl.updateMany(
      { phone: toKeepCustomer.phone },
      {
        $set: {
          phone: toArchiveCustomer.phone,
        },
      },
    );
    // update tickets with new customer id
    await ticketColl.updateMany(
      { customer: BSON.ObjectId(toArchiveCustomerId) },
      {
        $set: {
          customer: BSON.ObjectId(toKeepCustomerId),
        },
      },
    );
    // update the customer
    updatedCustomer = await customerColl.findOneAndUpdate(
      { _id: BSON.ObjectId(toKeepCustomerId) },
      {
        $set: {
          unread: toKeepCustomer.unread,
          created: toKeepCustomer.created,
          lastText: toKeepCustomer.lastText,
          lastInteraction: toKeepCustomer.lastInteraction,
          phone: toArchiveCustomer.phone,
        },
      },
      { returnNewDocument: true },
    );
    // Archive unused customer account
    await customerColl.findOneAndUpdate(
      { _id: BSON.ObjectId(toArchiveCustomerId) },
      {
        $set: {
          phone: `${toArchiveCustomer.phone}ARCHIVE`,
          unread: false,
          dontText: true,
          archived: true,
        },
      },
    );
  } catch (e) {
    console.error(e);
    throw e;
  }

  return updatedCustomer;
};
