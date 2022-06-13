const video = document.createElement("video");
const canvasElement = document.getElementById("qr-canvas");
const canvas = canvasElement.getContext("2d");
const container = document.getElementById("container")

let scanning = false;

qrcode.callback = (res) => {
   if (res) {
      var address = res.match(/nano_[a-z0-9]+/i);
      document.getElementById("finalAddress").value = address;
      var valid = validateNanoAddress();

      if (res.indexOf("amount") != -1) {
         var amountInRaw = res.match(/amount=(\d+)/i)[1];
         var amountInNano = nanocurrency.convert(amountInRaw, {from:"raw", to:"Nano"});
         document.getElementById("nanoAmount").value = amountInNano;
         autoFill(2);
      }
      scanning = false;

      video.srcObject.getTracks().forEach(track => {
         track.stop();
      });

      canvasElement.hidden = true;
      container.hidden = true;

      if (valid) {
         navigator.vibrate(100);
         showQR();
      }
   }
};

function requestCamera() {
   navigator.mediaDevices
      .getUserMedia({ video: { facingMode: "environment" } })
      .then(function(stream) {
      scanning = true;
      canvasElement.hidden = false;
      container.hidden = false;
      video.setAttribute("playsinline", true); // required to tell iOS safari we don't want fullscreen
      video.srcObject = stream;
      video.play();
      tick();
      scan();
    });
}

function tick() {
   canvasElement.height = video.videoHeight;
   canvasElement.width = video.videoWidth;
   canvas.drawImage(video, 0, 0, canvasElement.width, canvasElement.height);

   scanning && requestAnimationFrame(tick);
}

function scan() {
   try {
      qrcode.decode();
   } catch (e) {
      setTimeout(scan, 333);
   }
}
