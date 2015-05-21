
function register() {
  navigator.serviceWorker.register("sw.js", {scope: '/'}).then(
    function(serviceWorkerRegistration) {
      serviceWorkerRegistration.pushManager.subscribe().then(
        function(pushSubscription) {
          console.log(pushSubscription.endpoint);

          var data = new FormData();
          data.append('endpoint', pushSubscription.endpoint);
          var xhr = new XMLHttpRequest();
          xhr.open("POST", "/register");
          xhr.onload = function () {
            console.log(this.responseText);
            window.location = "/manage"
          };
          xhr.send(data);

        }, function(error) {
          console.log(error);
        }
      );
    }
  );
}

addEventListener("load", register, false);
