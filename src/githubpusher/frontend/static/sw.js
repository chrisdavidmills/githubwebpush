var gMessagePort = null;
self.addEventListener('message', function(event) {
    gMessagePort = event.ports[0];
});

this.onpush = function(event) {

    var currentdate = new Date(); 
    var datetime =    currentdate.getDate() + "/"
                    + currentdate.getMonth() + 1  + "/" 
                    + currentdate.getFullYear() + " @ "  
                    + currentdate.getHours() + ":"  
                    + currentdate.getMinutes() + ":" 
                    + currentdate.getSeconds();

  if (!gMessagePort) {
    return;
  }

  console.log(JSON.stringify(event.data))

  gMessagePort.postMessage({
            data: "Push from GitHubPusher! " + datetime + "(" + event.data + ")"
          });
}


