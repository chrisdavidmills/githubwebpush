

var gMessagePort = null;
self.addEventListener('message', function(event) {
    gMessagePort = event.ports[0];
});

function setupMessageChannel() {
  var messageChannel = new MessageChannel();
  messageChannel.port1.onmessage = function(event) {
    if (event.data.error) {
      console.log("manage.js: error="+event.data.error)
    } else {
      console.log("manage.js: data="+ JSON.stringify(event.data))
      location.reload(true);
    }
  }
  
  navigator.serviceWorker.ready.then(function(swr) { 
    if (!swr) {
      window.location = "/logout";
      return;
    }
    swr.active.postMessage("", [messageChannel.port2]);
  });
}

function logout() {
  navigator.serviceWorker.ready.then(function(serviceWorkerRegistration) {
    serviceWorkerRegistration.pushManager.getSubscription().then(
      function(subscription) {
        if (!subscription) {
          console.log("getRegistration() failed");
          window.location = "/logout";
          return;
        }
        subscription.unsubscribe();
      }
    );
  });
  window.location = "/logout";
}

addEventListener("load", setupMessageChannel, false);
