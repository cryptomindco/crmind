let x = document.querySelectorAll(".amount-number");
let btcAmountText = document.querySelectorAll(".btc-amount-number");
let dcrAmountText = document.querySelectorAll(".dcr-amount-number");
let ltcAmountText = document.querySelectorAll(".ltc-amount-number");
for (let i = 0, len = x.length; i < len; i++) {
  let num = formatToLocalString(Number(x[i].innerHTML), 2 , 2)
  x[i].innerHTML = num;
}

for (let i = 0, len = btcAmountText.length; i < len; i++) {
  let num = formatToLocalString(Number(btcAmountText[i].innerHTML), 8 , 8)
  btcAmountText[i].innerHTML = num;
}

for (let i = 0, len = dcrAmountText.length; i < len; i++) {
  let num = formatToLocalString(Number(dcrAmountText[i].innerHTML), 7 , 7)
  dcrAmountText[i].innerHTML = num;
}

for (let i = 0, len = ltcAmountText.length; i < len; i++) {
  let num = formatToLocalString(Number(ltcAmountText[i].innerHTML), 8, 8)
  ltcAmountText[i].innerHTML = num;
}

function shortenString(input, maxLength) {
  if(input.length <= maxLength) {
    return input
  }
  const firstLength = maxLength/2
  const firstStr = input.substring(0,firstLength - 1)
  const lastStr =  input.substring(input.length - firstLength, input.length)
  return firstStr + "..." + lastStr
}

function menuToggle() {
  const toggleMenu = document.querySelector(".menu");
  toggleMenu.classList.toggle("active");
}
$(document).on("click", function (e) {
  if ($(e.target).closest("#userToggle").length === 0) {
    if ($("#menuDropdown").hasClass("active")) {
      $("#menuDropdown").removeClass("active");
    }
  }
});

function formatDate (inputDate, format)  {
  if (!inputDate) return '';

  const padZero = (value) => (value < 10 ? `0${value}` : `${value}`);
  const parts = {
      yyyy: inputDate.getFullYear(),
      MM: padZero(inputDate.getMonth() + 1),
      dd: padZero(inputDate.getDate()),
      HH: padZero(inputDate.getHours()),
      hh: padZero(inputDate.getHours() > 12 ? inputDate.getHours() - 12 : inputDate.getHours()),
      mm: padZero(inputDate.getMinutes()),
      ss: padZero(inputDate.getSeconds()),
      tt: inputDate.getHours() < 12 ? 'AM' : 'PM'
  };

  return format.replace(/yyyy|MM|dd|HH|hh|mm|ss|tt/g, (match) => parts[match]);
}

function randomString(length, chars) {
  var mask = "";
  if (chars.indexOf("a") > -1) mask += "abcdefghijklmnopqrstuvwxyz";
  if (chars.indexOf("A") > -1) mask += "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  if (chars.indexOf("#") > -1) mask += "0123456789";
  var result = "";
  for (var i = length; i > 0; --i)
    result += mask[Math.floor(Math.random() * mask.length)];
  return result;
}

function toTransactionDetail(id) {
  window.location.href = "/transaction/detail?id=" + id;
}

function formatToLocalString(number, minFracDigit, maxFracDigit) {
  return number.toLocaleString('en-US', {
    minimumFractionDigits: minFracDigit,
    maximumFractionDigits: maxFracDigit
  })
}