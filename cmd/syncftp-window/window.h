#ifndef SYNCFTP_WINDOW_H
#define SYNCFTP_WINDOW_H

// Pencere olmadan çalışabilir mi (kütüphaneler + ekran). 1 = evet.
int window_probe(void);

// Pencereyi açar, kapanana kadar döner. smoke boş değilse sayfa açılınca durum smoke dosyasına yazılır ve çıkılır.
void window_run(const char *url, const char *title, const char *smoke);

// Sayfaya enjekte edilen ortak köprü (bkz. web/assets/titlebar.js)
#define SYNCFTP_BRIDGE_JS(NATIVE, POST) \
    "window.chrome=window.chrome||{};" \
    "window.chrome.webview={postMessage:function(m){" POST "}};" \
    "window.__SYNCFTP_FRAME={host:true,drag:true,native:'" NATIVE "'};"

#define SYNCFTP_SMOKE_JS \
    "addEventListener('load',function(){setTimeout(function(){" \
    "window.chrome.webview.postMessage('smoke:'+JSON.stringify({title:document.title,url:location.href," \
    "bar:!!document.getElementById('titlebar'),quick:!!document.querySelector('.tb-hizli')," \
    "page:(document.querySelector('#sayfa')||{}).textContent?true:false," \
    "storage:(function(){try{localStorage.setItem('x','1');return localStorage.getItem('x')==='1'}catch(e){return false}})()}))" \
    "},10000)});"

#endif
