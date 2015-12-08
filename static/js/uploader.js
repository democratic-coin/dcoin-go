function uploader(input, options) {
	var $this = this;

	// Default settings (mostly debug functions)
	this.settings = {
		prefix:'file',
		type:'file',
		autoUpload:false,
		url:window.location.href
	};
	$.extend(this.settings, options);

	this.input = input;
	this.xhr = new XMLHttpRequest();

	this.send = function(){
		// Make sure there is at least one file selected
		if($this.input.files.length < 1) {
			if($this.settings.error) $this.settings.error('Must select a file to upload');
			return false;
		}
		
		$this.singleSend($this.input.files[0]);
	};

	// Prep a single file for upload
	this.singleSend = function(file){
		var data = new FormData();
		data.append(String($this.settings.prefix),file);
		data.append('type',String($this.settings.type));
		$this.upload(data);
	};

	// The actual upload calls
	this.upload = function(data){
		$this.xhr.open('POST',$this.settings.url, true);
		$this.xhr.send(data);
	};

	// Modify options after instantiation
	this.setOpt = function(opt, val){
		$this.settings[opt] = val;
		return $this;
	};
	this.getOpt = function(opt){
		return $this.settings[opt];
	};

	// Set the input element after instantiation
	this.setInput = function(elem){
		$this.input = elem;
		return $this;
	};
	this.getInput = function(){
		return $this.input;
	};

	// Basic setup for the XHR stuff
	if(this.settings.progress) this.xhr.upload.addEventListener('progress',this.settings.progress,false);
	this.xhr.onreadystatechange = function(ev){
		if($this.xhr.readyState == 4) {
			if($this.xhr.status == 200) {
				if($this.settings.success) $this.settings.success(jQuery.parseJSON($this.xhr.responseText),ev);
				$this.input.value = '';
			}
			else {
				if($this.settings.error) $this.settings.error(ev);
			}
		}
	};

	// onChange event for autoUploads
	if(this.settings.autoUpload) this.input.onchange = this.send;
}