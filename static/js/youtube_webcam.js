delete window['YT'];
delete window['YTConfig'];

var YT = {loading: 0,loaded: 0};
var YTConfig = {'host': 'http://www.youtube.com'};
YT.loading = 1;(function(){var l = [];YT.ready = function(f) {if (YT.loaded) {f();} else {l.push(f);}};window.onYTReady = function() {YT.loaded = 1;for (var i = 0; i < l.length; i++) {try {l[i]();} catch (e) {}}};YT.setConfig = function(c) {for (var k in c) {if (c.hasOwnProperty(k)) {YTConfig[k] = c[k];}}};var a = document.createElement('script');a.id = 'www-widgetapi-script';a.src = 'https:' + '//s.ytimg.com/yts/jsbin/www-widgetapi-vfleeBgRM/www-widgetapi.js';a.async = true;var b = document.getElementsByTagName('script')[0];b.parentNode.insertBefore(a, b);})();

// 3. Define global variables for the widget and the player.
// The function loads the widget after the JavaScript code has
// downloaded and defines event handlers for callback notifications
// related to the widget.
var widget;
var player;
function onYouTubeIframeAPIReady() {
	widget = new YT.UploadWidget('widget', {
		width: 500,
		events: {
			'onUploadSuccess': onUploadSuccess,
			'onProcessingComplete': onProcessingComplete
		}
	});
}

// 4. This function is called when a video has been successfully uploaded.
function onUploadSuccess(event) {
	//alert('Video ID ' + event.data.videoId + ' was uploaded and is currently being processed. Please wait.');
    player = new YT.Player('player', {
        height: 390,
        width: 640,
        videoId: event.data.videoId,
        events: {}
    });

    $("#video_url_id").val(event.data.videoId);
    $("#refresh_youtube_div").css("display", "block");

    if (ytType!="promised_amount") {
        $.post('ajax?controllerName=saveVideo', {'video_url': 'youtu.be/' + event.data.videoId},
            function (data) {
            }, "json");
    }
}

// 5. This function is called when a video has been successfully processed.
function onProcessingComplete(event) {

}