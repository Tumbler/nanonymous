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
         var nano = thousandsRound(input * price);
         var afterTax = thousandsRound(CalculateTax(nano))
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
  if (Number.isInteger(num)) {
    return 0;
  }

  return num.toString().split('.')[1].length;
}
