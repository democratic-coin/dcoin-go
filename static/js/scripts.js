// JavaScript Document
function ResizeHeader(){
	'use strict';
	
	var w = $("body").width();
	if (w >= 976 && w < 1500) {
		$("header .logo").prependTo($("header .navbar-nav"));
	} else {
		$("header .logo").prependTo($("header"));
	}
	if (w < 976) {
		$("header .login").insertBefore($("header .navbar-nav ul"));
	} else {
		$("header .login").appendTo($(".mainmenu ul"));
	}
}

$(document).ready(function(){
	'use strict';
	
	jQuery.os = { name: (/(win|mac|linux|sunos|solaris|iphone|ipad)/.exec(navigator.platform.toLowerCase()) || [u])[0].replace('sunos', 'solaris') };
	if (jQuery.os.name === "mac" || jQuery.os.name === "iphone" || jQuery.os.name === "ipad") {
		$("body").addClass("macfix");
	}
	if (jQuery.os.name === "linux") {
		$("body").addClass("androidfix");
	}
});

$(window).load(function(){
	'use strict';
});

$(window).resize(function(){
	'use strict';
	
	ResizeHeader();
});

$(window).scroll(function(){
	'use strict';
});