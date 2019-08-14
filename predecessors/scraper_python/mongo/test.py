
from bson.code import Code
map = Code("function () {"
            "  this.tags.forEach(function(z) {"
            "    emit(z, 1);"
            "  });"
y            "}")
db.runners.mapReduce(

  // Map
  function(){
    for(var i =0; i < this.RunningSpeed.length; i++){
      var value={
        date: this.RunningSpeed[i].Date,
        speed: this.RunningSpeed[i].Value};

      // We emit all in single key value pairs
      emit(this.Name,value);
    }
  },

  // Reduce
  function(key,values){

    //In the beginning, delta is null
    var delta=0;

    // And we start to compare with the first value
    var last=values[0].speed;

    for(var idx=1; idx < values.length; idx++){

      // The absolute delta (all changed over all years)
      delta += Math.abs( values[idx].speed - last );

      // I think here was your problem.
      // You needed to save "last" years value for comparison with the current value.
      last=values[idx].speed;
    }

    var reduced = {

      // The relative delta (speed gain and loss over all years)
      delta_rel:values[values.length-1].speed - values[0].speed,

      delta_abs:delta
    };

    return reduced;

  },
  // Options
  {
     // Output collection
     out:"runner_test",
     // Sort order of the initial documents
     sort:{"RunningSpeed.Date":1}}
)