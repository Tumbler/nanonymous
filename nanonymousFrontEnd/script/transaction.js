function showQR() {
   var finalAddress = document.getElementById("finalAddress").value;
   document.getElementById("errorMessage").innerHTML = "";
   ajaxGetAddress(finalAddress);
}

function autoFill(caller) {
   switch(caller) {
      case 1:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("USDamount").value;
         var nano = thousandsRound(input / price);
         var afterTax = thousandsRound(CalculateTax(nano))
         document.getElementById("nanoAmount").value = nano;
         document.getElementById("afterTaxAmount").value = afterTax;
         break;
      case 2:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("nanoAmount").value;
         var usd = Math.round(input * price * 100) / 100;
         var afterTax = CalculateTax(parseFloat(input))
         document.getElementById("USDamount").value = usd;
         document.getElementById("afterTaxAmount").value = afterTax;
         break;
      case 3:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("afterTaxAmount").value;
         var nano = CalculateInverseTax(parseFloat(input))
         var usd = Math.round(nano * price * 100) / 100;
         document.getElementById("USDamount").value = usd;
         document.getElementById("nanoAmount").value = nano;
         break;
   }
}

function GetCurrentPrice() {
   fetch("https://api.coingecko.com/api/v3/simple/price?ids=nano&vs_currencies=usd").then(response => response.json()).then(data => SetCurrentPrice(data.nano.usd));
}

function SetCurrentPrice(data) {
   var price = data

   document.getElementById("nanoPrice").innerHTML = price;
}

// Basically just a 0.2% fee, but truncates any dust from the fee itself (but
// not from the payment so you can add your own dust if you so desire).
function CalculateTax(amount) {
   var feeWithDust = amount * 0.002;
   var fee = Math.floor(feeWithDust * 1000) / 1000;

   var finalVal = amount + fee;

   var precision = afterDecimal(amount);
   if (precision < 3) {
      precision = 3;
   }
   precision = 10 ** precision;

   return Math.round(finalVal * precision) / precision;
}

function CalculateInverseTax(amount) {
   var origWithDust = amount / 1.002;
   var origNoDust = Math.ceil(origWithDust * 1000) / 1000;

   var fee = thousandsRound(amount - origNoDust);
   var trueOrig = amount - fee;

   var precision = afterDecimal(amount);
   if (precision < 3) {
      precision = 3;
   }
   precision = 10 ** precision;

   return Math.round(trueOrig * precision) / precision;
}

function thousandsRound(number) {
   return Math.round(number * 1000) / 1000;
}

function afterDecimal(num) {
  if (isNaN(num) || Number.isInteger(num)) {
    return 0;
  }

  return num.toString().split('.')[1].length;
}

function validateNanoAddress() {
   var address = document.getElementById("finalAddress").value
   if (address.substr(0, 4) == "xrb_") {
      address = "nano_" + address.split("_")[1]
      document.getElementById("finalAddress").value = address
   }
   if (address.length != 65) {
      document.getElementById("errorMessage").innerHTML = "Address must be 65 characters long."
      document.getElementById("button").disabled = true
      return false
   } else if (!nanocurrency.checkAddress(address)) {
      document.getElementById("errorMessage").innerHTML = "Address invalid! Check for typos."
      document.getElementById("button").disabled = true
      return false
   } else {
      document.getElementById("errorMessage").innerHTML = ""
      document.getElementById("button").disabled = false
      return true
   }
}

async function ajaxGetAddress(finalAddress) {

   var req = new XMLHttpRequest();
   req.open("POST", "php/getNewAddress.php?address="+ finalAddress)

   var Nano = document.getElementById("afterTaxAmount").value;
   var raw = nanocurrency.convert(Nano, {from:"Nano", to:"raw"})

   // Wait for new address to come back from server and then display QR code.
   req.onload = function() {
      console.log(this.response);
      var reply = this.response.match(/address=(nano_[a-z0-9]+)/i);
      if (reply !== null && reply.length > 1) {
         address = reply[1];
         document.getElementById("TransactionInfo").innerHTML = "Tap QR code to open wallet if on mobile"

         var qrCodeText = "nano:" + address + "?amount=" + raw;

         document.getElementById("QRLink").href = qrCodeText;
         document.getElementById("qr-label").innerHTML = address;
         if (qrCodeText.length < 85) {
            var qrSize = 250;
         } else {
            var qrSize = 275;
         }
         var qr = new QRious({
            element: document.getElementById("QRCode"),
            foreground: '#151515', size: qrSize, padding: 5, value: qrCodeText
         });
         document.getElementById("QRdiv").hidden = false;
         document.getElementById("button").hidden = true;
         document.getElementById("button").style.display = "none";
         document.getElementById("scanQR").hidden = true;
         setTimeout(window.scrollTo(0,1000),100);
      } else {
         document.getElementById("errorMessage").innerHTML = "Something went wrong. Please try a different address or try again later.";

         // Don't connect to a transaction since one hasn't been started
         return
      }

      // Wait until transaction is complete and then post the hash.
      var req2 = new XMLHttpRequest();
      req2.open("POST", "php/getFinalHash.php?address="+ finalAddress, true)
      req2.timeout = 0; // No timeout

      req2.abort = function() {
         console.log("abort:", this.statusText, this.responseText);
      };
      req2.error = function() {
         console.log("error", this.statusText, this.responseText);
      };
      req2.timeout = function() {
         console.log("timeout", this.statusText, this.responseText);
      };
      req2.onabort = function() {
         console.log("onabort", this.statusText, this.responseText);
      };
      req2.onerror = function() {
         console.log("onerror", this.statusText, this.responseText);
      };
      req2.onprogress = function() {
         var line = this.responseText.match(/info=(.*)\n$/i);
         if (line !== null && line.length > 1) {
            if (line[1] == "amountTooLow") {
               document.getElementById("errorMessage").innerHTML = "The minimum transaction supported is 1 Nano. Your transaction has been refunded."
            } else if (line[1] == "") {
            }
         }
      };
      req2.ontimeout = function() {
         console.log("ontimeout", this.statusText, this.responseText);
      };

      req2.onload = function() {
         if (this.response.includes("hash=")) {
            console.log(this.response);
            var hash = this.response.match(/hash=([a-f0-9]+)/i)[1];

            document.getElementById("errorMessage").innerHTML = "";

            // Animate the address disappearing
            document.getElementById("payment-label").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("QRCode").classList.remove("animate-grow");
            document.getElementById("QRCode").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("qr-label").classList.remove("animate-fade-in");
            document.getElementById("qr-label").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("TransactionInfo").classList.add("animate-zipRight-out");
            document.getElementById("QRdiv").style.maxHeight = "0px";

            document.getElementById("HashLink").href = "https://www.nanolooker.com/block/" + hash;
            document.getElementById("HashLink").innerHTML = "Final hash:<br>" + hash;
            document.getElementById("HashLink").style.color = "#313133";
            document.getElementById("Hashdiv").style.maxHeight = "1000px";
            setTimeout(function(){ // delay by 900 ms
            document.getElementById("TransactionInfo").classList.remove("animate-zipRight-out");
            document.getElementById("TransactionInfo").style.textAlign = "center";
            document.getElementById("TransactionInfo").innerHTML = "Transaction Complete!"
            document.getElementById("TransactionInfo").classList.add("animate-zipRight-in");

            setTimeout(function(){ // delay by 100 ms


            document.getElementById("Hashdiv").classList.add("animate-zipRight-in");
            document.getElementById("HashLink").style.removeProperty("color");

            setTimeout(function(){ // delay by 1000 ms

            var confettiCanvas = document.createElement('canvas');
            confettiCanvas.style.position = 'fixed';
            confettiCanvas.style.width = '90%';
            confettiCanvas.style.height = '90%';
            confettiCanvas.style.top = '5%';
            confettiCanvas.style.left = '5%';
            confettiCanvas.style.zIndex = '-1';

            document.body.appendChild(confettiCanvas);

            myConfetti = confetti.create(confettiCanvas, {
               resize: true,
               useWorker: true
            });
            myConfetti({
               paricleCount: 80,
               spread: 140,
               startVelocity: 40,
               ticks: 175,
               origin: { y:0.7 }
            });
            setTimeout(() => {
               confetti.reset();
               document.body.removeChild(confettiCanvas);
            }, 5000);
            }, 1000);
            }, 100);
            }, 900);
            }, 100);
            }, 100);
            }, 100);
         } else {
            console.log(this.response);
            document.getElementById("errorMessage").innerHTML = "Something went wrong. Please try a different address or try again later.";
         }
      };
      req2.send();
   };
   req.send();

}

function sleep(ms) {
   return new Promise(resolve => setTimeout(resolve, ms));
}

function off() {
   var hash = "UwU";
            document.getElementById("payment-label").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("QRCode").classList.remove("animate-grow");
            document.getElementById("QRCode").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("qr-label").classList.remove("animate-fade-in");
            document.getElementById("qr-label").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("TransactionInfo").classList.add("animate-zipRight-out");
            document.getElementById("QRdiv").style.maxHeight = "0px";

            document.getElementById("HashLink").href = "https://www.nanolooker.com/block/" + hash;
            document.getElementById("HashLink").innerHTML = "Final hash:<br>" + hash;
            document.getElementById("HashLink").style.color = "#313133";
            document.getElementById("Hashdiv").style.maxHeight = "1000px";
            setTimeout(function(){ // delay by 900 ms
            document.getElementById("TransactionInfo").classList.remove("animate-zipRight-out");
            document.getElementById("TransactionInfo").style.textAlign = "center";
            document.getElementById("TransactionInfo").innerHTML = "Transaction Complete!"
            document.getElementById("TransactionInfo").classList.add("animate-zipRight-in");

            setTimeout(function(){ // delay by 100 ms


            document.getElementById("Hashdiv").classList.add("animate-zipRight-in");
            document.getElementById("HashLink").style.removeProperty("color");

            setTimeout(function(){ // delay by 1000 ms

            var confettiCanvas = document.createElement('canvas');
            confettiCanvas.style.position = 'fixed';
            confettiCanvas.style.width = '90%';
            confettiCanvas.style.height = '90%';
            confettiCanvas.style.top = '5%';
            confettiCanvas.style.left = '5%';
            confettiCanvas.style.zIndex = '-1';

            document.body.appendChild(confettiCanvas);

            myConfetti = confetti.create(confettiCanvas, {
               resize: true,
               useWorker: true
            });
            myConfetti({
               paricleCount: 80,
               spread: 140,
               startVelocity: 40,
               ticks: 175,
               origin: { y:0.6 }
            });
            setTimeout(() => {
               confetti.reset();
               document.body.removeChild(confettiCanvas);
            }, 5000);
            }, 1500);
            }, 100);
            }, 900);
            }, 100);
            }, 100);
            }, 100);

   document.getElementById("reset").onclick = on;
}

function on() {
   console.log("on")
   document.getElementById("Hashdiv").style.maxHeight = "0px";
   document.getElementById("QRdiv").style.maxHeight = "1000px";

      document.getElementById("QRCode").classList.remove("animate-zipRight-out");
      document.getElementById("QRCode").classList.add("animate-grow");
      document.getElementById("qr-label").classList.remove("animate-zipRight-out");
      document.getElementById("qr-label").classList.add("animate-fade-in");
      document.getElementById("payment-label").classList.remove("animate-zipRight-out");

   document.getElementById("QRdiv").hidden = false;
   document.getElementById("reset").onclick = off;
}

function copyAddress() {
   var label = document.getElementById("qr-label");
   var tooltip = document.getElementById("tooltip");
   var text = label.innerHTML;

   navigator.clipboard.writeText(text);

   tooltip.style.opacity = '1';

   setTimeout(function(){tooltip.style.opacity = '0';}, 3000);
}
