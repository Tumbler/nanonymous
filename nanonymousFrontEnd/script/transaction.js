let nanonymousFee = 0.01;
let QRactive = false;
let middleAddress = "";
let beta = false;

let mobileOrTablet = mobileOrTabletCheck()

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
         var afterTax = thousandsRound(CalculateInverseTax(nano))
         document.getElementById("nanoAmount").value = nano;
         document.getElementById("afterTaxAmount").value = afterTax;
         break;
      case 2:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("nanoAmount").value;
         var usd = Math.round(input * price * 100) / 100;
         var afterTax = CalculateInverseTax(parseFloat(input))
         if (!isNaN(usd) && !isNaN(afterTax)) {
            document.getElementById("USDamount").value = usd;
            document.getElementById("afterTaxAmount").value = afterTax;
         } else {
            document.getElementById("USDamount").value = "";
            document.getElementById("afterTaxAmount").value = "";
         }
         break;
      case 3:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("afterTaxAmount").value;
         var nano = CalculateTax(parseFloat(input))
         var usd = Math.round(nano * price * 100) / 100;
         if (!isNaN(usd) && !isNaN(nano)) {
            document.getElementById("USDamount").value = usd;
            document.getElementById("nanoAmount").value = nano;
         } else {
            document.getElementById("USDamount").value = "";
            document.getElementById("nanoAmount").value = "";
         }
         break;
   }

   if (QRactive) {
      let Nano = document.getElementById("afterTaxAmount").value;
      let raw = nanocurrency.convert(Nano, {from:"Nano", to:"raw"})
      if (isNaN(raw)) {
         var qrCodeText = "nano:" + middleAddress;
      } else {
         var qrCodeText = "nano:" + middleAddress + "?amount=" + raw;
      }
      if (qrCodeText.length < 85) {
         var qrSize = 260;
      } else {
         var qrSize = 275;
      }

      document.getElementById("QRLink").href = qrCodeText;

      var qr = new QRious({
         element: document.getElementById("QRCode"),
         foreground: '#151515', size: qrSize, padding: 5, value: qrCodeText
      });
   }
}

function GetCurrentPrice() {
   fetch("https://api.coingecko.com/api/v3/simple/price?ids=nano&vs_currencies=usd").then(response => response.json()).then(data => SetCurrentPrice(data.nano.usd));
}

let nanoVal = 0
function SetCurrentPrice(data) {
   nanoVal = data

   document.getElementById("nanoPrice").innerHTML = nanoVal.toFixed(6);
}

function GetCurrentFee() {

   var req = new XMLHttpRequest();
   req.open("POST", "php/getFee.php")

   req.onload = function() {
      console.log(this.response);
      var reply = this.response.match(/fee=([0-9]+\.[0-9]+)/i);
      if (reply !== null && reply.length > 1) {
         nanonymousFeePercent = parseFloat(reply[1]);
         nanonymousFee = nanonymousFeePercent / 100
         if (nanonymousFeePercent == 0) {
            document.getElementById("nanonymousFee").innerHTML = "Free!";
            document.getElementById("afterFeeRow").hidden = true;
            beta = true
            document.getElementById("calculator").deleteCaption();
         } else {
            document.getElementById("nanonymousFee").innerHTML = nanonymousFeePercent.toString().concat("% (less than a percent)");
         }
      } else {
         document.getElementById("errorMessage").innerHTML = "Can't contact our servers right now. Please try again later.";
         document.getElementById("errorMessage").scrollIntoView();
         document.getElementById("addressInputContainer").hidden = true;
         document.getElementById("button").hidden = true;
      }
   }
   req.send();
}

// Basically just a 0.2% fee, but truncates any dust from the fee itself (but
// not from the payment so you can add your own dust if you so desire).
function CalculateTax(amount) {
   var feeWithDust = amount * nanonymousFee;
   var fee = Math.floor(feeWithDust * 1000) / 1000;

   var finalVal = amount - fee;

   var precision = afterDecimal(amount);
   if (precision < 3) {
      precision = 3;
   }
   precision = 10 ** precision;

   if (amount < 1 && !beta) {
      document.getElementById("errorMessage").innerHTML = "The minimum transaction supported is 1 Nano.";
   } else {
      document.getElementById("errorMessage").innerHTML = "";
   }

   return Math.round(finalVal * precision) / precision;
}

function CalculateInverseTax(amount) {
   var origWithDust = amount / (1 - nanonymousFee);
   var fee = Math.floor((origWithDust - amount) * 1000) / 1000;

   var trueOrig = amount + fee;

   var precision = afterDecimal(amount);
   if (precision < 3) {
      precision = 3;
   }
   precision = 10 ** precision;

   if (trueOrig < 1 && !beta) {
      document.getElementById("errorMessage").innerHTML = "The minimum transaction supported is 1 Nano.";
   } else {
      document.getElementById("errorMessage").innerHTML = "";
   }

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

   let url = "php/getNewAddress.php?address="+ finalAddress;

   if (document.getElementById("advancedCheck").checked) {
      let numSends = document.getElementById("numSends").value;

      if (numSends > 1) {
         url += "&percents="
         for (let i = 1; i <= numSends; i++) {
            url += document.getElementById("send"+ i +"Percent").value + ",";
         }
         // Remove trailing comma
         url = url.slice(0, -1);
      }

      if (document.getElementById("delayCheck").checked) {
         url += "&delays="
         if (numSends == 1) {
            let seconds = parseInt(document.getElementById("send0Min").value) * 60
            seconds += parseInt(document.getElementById("send0Sec").value)
            url += seconds
         } else {
            for (let i = 1; i <= numSends; i++) {
               let seconds = parseInt(document.getElementById("send"+ i +"Min").value) * 60
               seconds += parseInt(document.getElementById("send"+ i +"Sec").value)
               url += seconds + ",";
            }
            // Remove trailing comma
            url = url.slice(0, -1);
         }
      }
   }

   console.log("url="+ url);

   var req = new XMLHttpRequest();
   req.open("POST", "php/getNewAddress.php?address="+ finalAddress)

   var Nano = document.getElementById("afterTaxAmount").value;
   var raw = nanocurrency.convert(Nano, {from:"Nano", to:"raw"})

   // Wait for new address to come back from server and then display QR code.
   req.onload = function() {
      console.log(this.response);
      var reply = this.response.match(/address=(nano_[a-z0-9]+)/i);
      var info = this.response.match(/info=(.*)\n$/i);
      var bridge = this.response.match(/bridge=(\w+)/i);
      if (reply !== null && reply.length > 1) {
         middleAddress = reply[1];

         if (isNaN(raw)) {
            var qrCodeText = "nano:" + middleAddress;
         } else {
            var qrCodeText = "nano:" + middleAddress + "?amount=" + raw;
         }

         document.getElementById("QRLink").href = qrCodeText;
         document.getElementById("qr-label").innerHTML = middleAddress.concat("<img src=\"images/copyWhite.png\" style=\"width:17px;height:18px;padding:0px 0px 10px 5px\">");
         if (qrCodeText.length < 85) {
            var qrSize = 260;
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
         setTimeout(window.scrollTo(0, document.body.scrollHeight),100);
         QRactive = true;

         document.getElementById("finalAddress").disabled = true;
         document.getElementById("advancedCheck").disabled = true;
         let nodes = document.getElementById("advancedOptions").getElementsByTagName('*');
         for (let i = 0; i < nodes.length; i++) {
            nodes[i].disabled = true
         }

         if (bridge !== null && bridge.length > 1 && bridge[1] == "true") {
            document.getElementById("updateMessage").innerHTML = "Your recipient is also using Nanonymous. Your transaction will go to their final address, but you won't receive a final hash to respect their privacy. (The fee will only be applied once)";
            document.getElementById("updateMessage").hidden = false;
         }

         if (mobileOrTablet) {
            var tooltiptap = document.getElementById("tooltiptap");

            setTimeout(function(){
               tooltiptap.style.opacity = '1';
               setTimeout(function(){
                  tooltiptap.style.opacity = '0';
               }, 3000);
            }, 1500);
         }
      } else if (info !== null && info.length > 1) {
         document.getElementById("errorMessage").innerHTML = info[1];
         document.getElementById("errorMessage").scrollIntoView();
      } else {
         document.getElementById("errorMessage").innerHTML = "Something went wrong. Please try a different address or try again later.";
         document.getElementById("errorMessage").scrollIntoView();

         // Don't connect to a transaction since one hasn't been started
         return
      }

      // Wait until transaction is complete and then post the hash.
      var req2 = new XMLHttpRequest();
      req2.open("POST", "php/getFinalHash.php?address="+ middleAddress, true)
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
         var update = this.responseText.match(/update=(.*)\n$/i);
         if (line !== null && line.length > 1) {
            console.log(this.responseText);
            document.getElementById("errorMessage").innerHTML = line[1]
            document.getElementById("errorMessage").scrollIntoView();
            document.getElementById("updateMessage").hidden = true;
            if (this.response.includes("timed out")) {
               req2.abort();
            }
         } else if (update !== null && update.length > 1) {
            console.log(this.responseText);
            document.getElementById("updateMessage").innerHTML = update[1];
            document.getElementById("updateMessage").hidden = false;
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
            document.getElementById("errorMessage").scrollIntoView();
            document.getElementById("updateMessage").hidden = true;

            // Animate the address disappearing
            document.getElementById("payment-label").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("QRCode").classList.remove("animate-grow");
            document.getElementById("QRCode").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("qr-label").classList.remove("animate-fade-in");
            document.getElementById("qr-label").classList.add("animate-zipRight-out");
            setTimeout(function(){ // delay by 100 ms
            document.getElementById("QRdiv").style.maxHeight = "0px";

            if (hash.length < 32) {
               document.getElementById("HashLink").innerHTML = "";
               document.getElementById("HashLink").style.color = "#313133";
            } else {
               document.getElementById("HashLink").href = "https://www.nanolooker.com/block/" + hash;
               document.getElementById("HashLink").target = "blank";
               document.getElementById("HashLink").innerHTML = "Final hash:<br>" + hash;
               document.getElementById("HashLink").style.color = "#313133";
            }
            document.getElementById("Hashdiv").style.maxHeight = "1000px";
            document.getElementById("tooltiptap").hidden = true;

            setTimeout(function(){ // delay by 900 ms
            document.getElementById("TransactionInfo").classList.remove("animate-zipRight-out");
            document.getElementById("TransactionInfo").style.textAlign = "center";
            document.getElementById("TransactionInfo").innerHTML = "<b>Transaction Complete!</b>"
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

            // Find the y-percent where the final hash is and put confetti there.
            var y = document.getElementById("HashLink").getBoundingClientRect().y;
            var percentY = y/window.innerHeight;
            myConfetti({
               paricleCount: 80,
               spread: 140,
               startVelocity: 40,
               ticks: 175,
               origin: { y: percentY }
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

            QRactive = true;
            req2.abort();
         } else {
            console.log(this.response);
            document.getElementById("errorMessage").innerHTML = "Something went wrong. Please try a different address or try again later.";
            document.getElementById("errorMessage").scrollIntoView();
         }
      };
      req2.send();
   };
   req.send();

}

function copyAddress() {
   var label = document.getElementById("qr-label");
   var tooltip = document.getElementById("tooltip");
   var text = label.textContent;

   navigator.clipboard.writeText(text);

   tooltip.style.opacity = '1';

   setTimeout(function(){tooltip.style.opacity = '0';}, 3000);
}

function changeCurrency() {
   let info = document.getElementById("currencyDropDown").value.split(",")
   let newCurrency = ""
   if (info !== null && info.length > 1) {
      newCurrency = info[1]
   }

   let req = new XMLHttpRequest();
   req.open("POST", "php/getCurrencyValue.php?curr="+ newCurrency)
   req.onload = function() {
      console.log(this.response);
      var reply = this.response.match(/val=([0-9]+(?:\.[0-9]+)?),([0-9]+(?:\.[0-9]+)?)/i);
      if (reply !== null && reply.length > 2) {
         usdVal = parseFloat(reply[1]);
         curVal = parseFloat(reply[2]);

         // The API I'm using is based in euros, so I have to do an extra
         // calculation to find the nano to curr val.
         newVal = nanoVal * curVal / usdVal
         document.getElementById("nanoPrice").innerHTML = newVal.toFixed(6);

         if (info !== null && info.length > 1) {
            // Change all the currency symbols
            let labels = document.getElementsByClassName('currSym');
            [].slice.call(labels).forEach(function(label) {
               label.innerHTML = info[0];
            });
         }

         if (document.getElementById("USDamount").value.length > 0) {
            autoFill(1);
         }
      }
   }
   req.send();
}

function timeCheck(textBox) {
   if (textBox.value > 59) {
      textBox.value = 59;
   } else if (textBox.value < 0) {
      textBox.value = 0;
   }
}

function balancePercents(textBox) {
   if (textBox.value > 99) {
      textBox.value = 99;
   } else if (textBox.value < 1 && textBox.value.length > 0) {
      textBox.value = 1;
   }

   let numSends = document.getElementById("numSends").value;

   for (let i = numSends; i > 0; i--) {
      thisBox = document.getElementById("send"+ i +"Percent");
      if (thisBox === textBox) {
         // Don't alter the box we're editing.
         continue;
      }

      let changeNeeded = 100 - getTotalPercent();

      thisBox.value = parseInt(thisBox.value) + changeNeeded

      if (thisBox.value > 99) {
         thisBox.value = 99;
      } else if (thisBox.value < 1) {
         thisBox.value = 1;
      } else {
         // No more change needed
         break
      }
   }

   if (getTotalPercent() != 100) {
      document.getElementById("errorMessage").innerHTML = "Percents do not add up to 100%.";
   } else {
      document.getElementById("errorMessage").innerHTML = "";
   }
}

function getTotalPercent() {
   let numSends = document.getElementById("numSends").value;
   let totalPercent = 0
   for (let i = 1; i <= numSends; i++) {
      num = parseInt(document.getElementById("send"+ i +"Percent").value);
      if (!isNaN(num)) {
         totalPercent += num
      }
   }

   return totalPercent
}


function toggleOptions() {
   let checked = document.getElementById("advancedCheck").checked;

   if (checked) {
      document.getElementById("advancedOptions").hidden = false;
      document.getElementById("numSends").value = 2;
      changeSends();
   } else {
      document.getElementById("advancedOptions").hidden = true;
   }
}

function changeSends() {
   let numSends = document.getElementById("numSends").value;
   if (numSends < 1) {
      numSends = 1;
      // Let them erase
      if (document.getElementById("numSends").value.length > 0) {
         document.getElementById("numSends").value = 1
      }
   } else if (numSends > 5) {
      numSends = 5;
      document.getElementById("numSends").value = 5
   }

   if (numSends > 1) {
      document.getElementById("delayInput0").hidden = true;
   } else if (document.getElementById("delayCheck").checked) {
      document.getElementById("delayInput0").hidden = false;
   }

   if (numSends < 2) {
      document.getElementById("send1").style.display = "none";
      document.getElementById("send2").style.display = "none";
   } else {
      document.getElementById("send1").style.display = "block";
      document.getElementById("send2").style.display = "block";
   }

   if (numSends < 3) {
      document.getElementById("send3").style.display = "none";
   } else {
      document.getElementById("send3").style.display = "block";
   }

   if (numSends < 4) {
      document.getElementById("send4").style.display = "none";
   } else {
      document.getElementById("send4").style.display = "block";
   }

   if (numSends < 5) {
      document.getElementById("send5").style.display = "none";
   } else {
      document.getElementById("send5").style.display = "block";
   }

   // Randomize percents
   if (numSends > 1) {
      let percentLeft = 100
      let randomPercents = new Array(numSends);
      for (let i = 0; i < numSends; i++) {
         if (i == numSends -1) {
            // Last one; soak up the rest
            randomPercents[i] = percentLeft;
         } else {
            value = getRandomInt(1, .8 * percentLeft);
            randomPercents[i] = value
            percentLeft -= value
         }
      }

      shuffle(randomPercents)
      document.getElementById("send1Percent").value = randomPercents[0]
      if (randomPercents.length > 1) {
         document.getElementById("send2Percent").value = randomPercents[1]
      }
      if (randomPercents.length > 2) {
         document.getElementById("send3Percent").value = randomPercents[2]
      }
      if (randomPercents.length > 3) {
         document.getElementById("send4Percent").value = randomPercents[3]
      }
      if (randomPercents.length > 4) {
         document.getElementById("send5Percent").value = randomPercents[4]
      }
   }
}

function toggleDelays() {
   let displayBool = document.getElementById("delayCheck").checked;

   if (document.getElementById("numSends").value == 1) {
      document.getElementById("delayInput0").hidden = !displayBool;

   } else {
      document.getElementById("delayInput0").hidden = true;
   }

   let delayInputs = document.getElementsByClassName('delayInput');
   [].slice.call(delayInputs).forEach(function(delayInput) {
      delayInput.hidden = !displayBool;
   });

   // Randomize delays
   if (displayBool) {
      document.getElementById("send0Min").value = getRandomInt(0, 14)
      document.getElementById("send0Sec").value = getRandomInt(0, 59)
      document.getElementById("send1Min").value = getRandomInt(0, 14)
      document.getElementById("send1Sec").value = getRandomInt(0, 59)
      document.getElementById("send2Min").value = getRandomInt(0, 14)
      document.getElementById("send2Sec").value = getRandomInt(0, 59)
      document.getElementById("send3Min").value = getRandomInt(0, 14)
      document.getElementById("send3Sec").value = getRandomInt(0, 59)
      document.getElementById("send4Min").value = getRandomInt(0, 14)
      document.getElementById("send4Sec").value = getRandomInt(0, 59)
      document.getElementById("send5Min").value = getRandomInt(0, 14)
      document.getElementById("send5Sec").value = getRandomInt(0, 59)
   }
}

function mobileOrTabletCheck() {
  let check = false;
  (function(a){if(/(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|iris|kindle|lge |maemo|midp|mmp|mobile.+firefox|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows ce|xda|xiino|android|ipad|playbook|silk/i.test(a)||/1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-/i.test(a.substr(0,4))) check = true;})(navigator.userAgent||navigator.vendor||window.opera);
  return check;
};

function getRandomInt(min, max) {
   const randomBuffer = new Uint32Array(1);

   window.crypto.getRandomValues(randomBuffer);

   let randomNumber = randomBuffer[0] / (0xffffffff + 1);

   min = Math.ceil(min);
   max = Math.floor(max);
   return Math.floor(randomNumber * (max - min + 1)) + min;
}

function shuffle(array) {
  let currentIndex = array.length,  randomIndex;

  // While there remain elements to shuffle.
  while (currentIndex > 0) {

    // Pick a remaining element.
    randomIndex = Math.floor(Math.random() * currentIndex);
    currentIndex--;

    // And swap it with the current element.
    [array[currentIndex], array[randomIndex]] = [
      array[randomIndex], array[currentIndex]];
  }

  return array;
}
