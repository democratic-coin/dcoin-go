onmessage = function (obj) {
    importScripts("/static/js/md5.js");

    importScripts("/static/js/asn1hex-1.1.min.js");
    importScripts("/static/js/jsbn.js");
    importScripts("/static/js/jsbn2.js");
    importScripts("/static/js/base64.js");
    importScripts("/static/js/sha1.js");
    importScripts("/static/js/rsa.js");
    importScripts("/static/js/rsa2.js");
    importScripts("/static/js/rsasign-1.2.min.js");
    importScripts("/static/js/rsapem-1.1.js");

    importScripts("/static/js/hex2a.js");
    importScripts("/static/js/aes.js");
    importScripts("/static/js/enc-base64-min.js");
    importScripts("/static/js/crypto-js-aes.js");
    importScripts("/static/js/crypto-js-mode-ecb.js");
    importScripts("/static/js/crypto-js-pad-iso10126.js");

    var hSig;
    var modulus;
    var exp;
    var decrypt_PEM;
    var error;
    var key = obj.data.key;
    var pass = obj.data.pass;

   // console.log(key);
  //  console.log(pass);
    key = key.trim();
    // ключ может быть незашифрованным, но без BEGIN RSA PRIVATE KEY
    if (key.substr(0,4) == 'MIIE')
        decrypt_PEM = '-----BEGIN RSA PRIVATE KEY-----'+key+'-----END RSA PRIVATE KEY-----';
    else if (pass && key.indexOf('RSA PRIVATE KEY')==-1) {
        try{
            ivAndText = atob(key);
            iv = ivAndText.substr(0, 16);
            encText = ivAndText.substr(16);
            cipherParams = CryptoJS.lib.CipherParams.create({
                ciphertext: CryptoJS.enc.Base64.parse(btoa(encText))
            });

            pass = CryptoJS.enc.Latin1.parse(hex_md5(pass));
            var decrypted = CryptoJS.AES.decrypt(cipherParams, pass, {mode: CryptoJS.mode.CBC, iv: CryptoJS.enc.Utf8.parse(iv), padding: CryptoJS.pad.Iso10126 });
            decrypt_PEM = hex2a(decrypted.toString());
        } catch(e) {
            decrypt_PEM = 'invalid decrypt ('+e+')';
        }
    } else {
        decrypt_PEM = key
    }

    if (typeof decrypt_PEM != "string" || decrypt_PEM.indexOf('RSA PRIVATE KEY') == -1) {
        error = "incorrect_key (size="+decrypt_PEM.length+")";
        if (decrypt_PEM.length < 100) {
            error+=decrypt_PEM
        }
    } else {
        var rsa = new RSAKey();
        var a = rsa.readPrivateKeyFromPEMString(decrypt_PEM);
        modulus = a[1];
        exp = a[2];
        if (obj.data.forsign != "") {
            //console.log(obj.data.forsign)
            hSig = rsa.signString(obj.data.forsign, 'sha1');
            //console.log(hSig)
        }
        delete rsa;
    }

    var result =
    {
        hSig: hSig,
        modulus: modulus,
        exp: exp,
        decrypt_key: decrypt_PEM,
        error: error
    };
  //  console.log("result", result);

    postMessage(result);
}