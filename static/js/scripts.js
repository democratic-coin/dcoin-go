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

function HideMenu(){
	'use strict';
	
	if ($("header").hasClass("on")) {
		$("header").removeClass("on");
	}
}