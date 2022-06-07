// TODO w3schools modal image

function showQR() {
   var textAddress = document.getElementById("finalAddress").value;
   document.getElementById("QRCode").innerHTML = textAddress

   var qr = new QRious({
      element: document.getElementById("QRImage"),
      size: 250, value: textAddress
   });
}

function autoFill(caller) {
   switch(caller) {
      case 1:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("USDamount").value;
         var nano = Math.round(input * price * 1000) / 1000;
         var afterTax = CalculateTax(nano)
         document.getElementById("nanoAmount").value = nano;
         document.getElementById("afterTaxAmount").value = afterTax;
         break;
      case 2:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("nanoAmount").value;
         var usd = Math.round(input / price * 100) / 100;
         var afterTax = CalculateTax(parseFloat(input))
         document.getElementById("USDamount").value = usd;
         document.getElementById("afterTaxAmount").value = afterTax;
         break;
      case 3:
         var price = document.getElementById("nanoPrice").innerHTML;
         var input = document.getElementById("afterTaxAmount").value;
         var nano = CalculateInverseTax(parseFloat(input))
         var usd = Math.round(nano / price * 100) / 100;
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

   console.log(price);
}

// Adds .001 for every .5 nano converted (Essentially a .2% tax, but without all
// the dust)
function CalculateTax(amount) {
   var div = amount / .5;
   var truncated = Math.floor(div);

   // Do weird amplified stuff because floating point values are literally the worst
   var amplified = amount * 1000 + truncated;
   var finalVal = Math.round(amplified) / 1000;

   return finalVal;
}

function CalculateInverseTax(amount) {
   var div = amount / .5;
   var truncated = Math.floor(div);

   // Do weird amplified stuff because floating point values are literally the worst
   var amplified = amount * 1000 - truncated;
   var finalVal = Math.round(amplified) / 1000;

   // The inverse function isn't exact, so backwards check the answer and reverse solve if needed.
   var check = CalculateTax(finalVal);
   if (check == amount) {
      console.log("yes")
      return finalVal;
   } else {
      var diff = amount - check;
      return Math.round((finalVal + diff) * 1000) /1000
   }
}
