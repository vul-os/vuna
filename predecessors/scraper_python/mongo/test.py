
addr = "192.168.8.120"
port = 27017
from bson.code import Code
import pymongo
conn = pymongo.MongoClient(host="192.168.8.120:27017")
db = conn.trophyseeds
collection = db.data

a = collection.find().sort('date', -1)

b = collection.aggregate([{
        "$group": {
            "_id": {
              "name": "$name",
            },
            "entries": {
              "$push": {
                "date": "$date",
                "stock": "$stock",
              }
            }
        },
        "$group": {
          "_id": {
            "name": "$name",
          },
          "stdDev": {
            "$stdDevPop": "$stock"
          },
          "maxValue": {"$max": "$stock"},
          "minValue": {"$min": "$stock"},
          "firstValue": {"$first": "$stock"},
          "lastValue": {"$last": "$stock"}
          # "entries": {
          #   "$push": {
          #     "stock": "$stock"
          #   }
          # }
        },
},{"$sort": {"stdDev": -1}}
])

# c = b.sort("stdDev", -1)



for c in b:
  print(c)

#
# map = Code("function() { emit(this.name,this.stock);};")
# reduce = Code(""
#               ""
#               )






# for c in col:
#   print(c)
# from bson.code import Code
# map = Code("function () {"
#             "  this.tags.forEach(function(z) {"
#             "    emit(z, 1);"
#             "  });"
# y            "}")
# db.runners.mapReduce(
#
#   // Map
#   function(){
#     for(var i =0; i < this.RunningSpeed.length; i++){
#       var value={
#         date: this.RunningSpeed[i].Date,
#         speed: this.RunningSpeed[i].Value};
#
#       // We emit all in single key value pairs
#       emit(this.Name,value);
#     }
#   },
#
#   // Reduce
#   function(key,values){
#
#     //In the beginning, delta is null
#     var delta=0;
#
#     // And we start to compare with the first value
#     var last=values[0].speed;
#
#     for(var idx=1; idx < values.length; idx++){
#
#       // The absolute delta (all changed over all years)
#       delta += Math.abs( values[idx].speed - last );
#
#       // I think here was your problem.
#       // You needed to save "last" years value for comparison with the current value.
#       last=values[idx].speed;
#     }
#
#     var reduced = {
#
#       // The relative delta (speed gain and loss over all years)
#       delta_rel:values[values.length-1].speed - values[0].speed,
#
#       delta_abs:delta
#     };
#
#     return reduced;
#
#   },
#   // Options
#   {
#      // Output collection
#      out:"runner_test",
#      // Sort order of the initial documents
#      sort:{"RunningSpeed.Date":1}}
# )