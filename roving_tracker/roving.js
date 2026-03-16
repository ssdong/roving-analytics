(function(){
  'use strict';

  var loc = window.location;
  var doc = window.document;
  var scriptEl = doc.currentScript;

  var endpoint = scriptEl.getAttribute('data-api') || (new URL(scriptEl.src).origin + '/api/event');

  if (/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(loc.hostname)) return;

  // https://sentry.io/answers/generate-random-string-characters-in-javascript
  function createRandomString(length) {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let result = "";
    const randomArray = new Uint8Array(length);
    crypto.getRandomValues(randomArray);
    randomArray.forEach((number) => {
      result += chars[number % chars.length];
    });
    return result;
  }

  function getSessionSalt() {
    var saltKey = 'roving_session_salt';
    try {
      // Check if we already have a salt for this tab session
      var salt = window.sessionStorage.getItem(saltKey);
      if (!salt) {
        salt = createRandomString(32);
        window.sessionStorage.setItem(saltKey, salt);
      }
      return salt;
    } catch (e) {
      // Fallback for if interacting with session storage failed
      return createRandomString(32);
    }
  }

  var payload = JSON.stringify({
    e: 'pageview',
    u: loc.href,
    r: doc.referrer || null,
    t: Date.now(),
    s: getSessionSalt()
  });

  // Send data
  if (navigator.sendBeacon) {
    navigator.sendBeacon(endpoint, payload);
  } else {
    var req = new XMLHttpRequest();
    req.open('POST', endpoint, true);
    req.setRequestHeader('Content-Type', 'text/plain');
    req.send(payload);
  }
})();