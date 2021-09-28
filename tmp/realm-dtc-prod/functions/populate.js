/*
  This function is run when a GraphQL Query is made requesting your
  custom field name. The return value of this function is used to
  populate the resolver generated from your Payload Type.

  This function expects the following input object:

  {
    "type": "object",
    "title": "populateInput",
    "required": [
      "collection",
      "sourceField"
    ],
    "properties": {
      "limit": {
        "bsonType": "int"
      },
      "skip": {
        "bsonType": "int"
      },
      "sort": {
        "bsonType": "int"
      },
      "collection": {
        "bsonType": "string"
      },
      "sourceField": {
        "bsonType": "string"
      },
      "targetField": {
        "bsonType": "string"
      }
    }
  }

  And the following payload object:

  {
    "type": "object",
    "title": "populateResult",
    "properties": {
      "resultType": {
        "bsonType": "string"
      },
      "populated": {
        "bsonType": "array",
        "items": {
          "bsonType": "object"
        }
      }
    }
  }
*/

exports = async (input, doc) => {
  const skip = input.skip || 0;
  const limit = input.limit || 1000; // TODO: reasonable max? Should this even be limited?
  const sort = input.sort || 1;
  const targetField = input.targetField || '_id';

  const cluster = context.services.get('mongodb-atlas');
  const targetColl = cluster.db('verbenergy').collection(input.collection);

  let toPopulate = doc[input.sourceField];
  if (sort === -1) {
    const reversed = []; // TODO: Figure out why array.reverse wasn't working
    for (let i = toPopulate.length - 1; i >= 0; i--) {
      reversed.push(toPopulate[i]);
    }
    toPopulate = reversed;
  }

  const populated = await Promise.all(toPopulate.slice(skip, skip + limit).map(async (val) => {
    const query = {};
    query[targetField] = val;
    return targetColl.findOne(query);
  }));
  return populated.filter((entry) => entry);
};
