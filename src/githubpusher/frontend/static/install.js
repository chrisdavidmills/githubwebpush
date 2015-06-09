
function register() {
  navigator.serviceWorker.register("sw.js", {scope: '/'});

  navigator.serviceWorker.ready.then(
    function(swr) {

      swr.pushManager.subscribe().then(
        function(pushSubscription) {

          var data = new FormData();
          data.append('endpoint', pushSubscription.endpoint);
          data.append('subscription', pushSubscription.subscriptionId);

          var xhr = new XMLHttpRequest();
          xhr.open("POST", "/register");
          xhr.onload = function () {
            window.location = "/manage"
          };
          xhr.send(data);

        },

        function(error) {
          console.log(error);
        }
      );
    }
  );
}

addEventListener("load", register, false);
