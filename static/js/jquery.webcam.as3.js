/**
* jQuery AS3 Webcam
*
* Copyright (c) 2012, Sergey Shilko (sergey.shilko@gmail.com)
*
* @author Sergey Shilko
* @see https://github.com/sshilko/jQuery-AS3-Webcam
*
**/
try {
$(document).ready( function() {
    var webcam = {
        previewWidth: 320,
        previewHeight: 240,

        resolutionWidth: 320,
        resolutionHeight: 240,

        videoDeblocking: 0,
        videoSmoothing: 0,

        /**
         * Determine if we need to stretch or scale the captured stream
         *
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/Stage.html#scaleMode
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/StageScaleMode.html
         */
        StageScaleMode: 'exactFit', //

        /**
         * Aligns video output on stage
         *
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/StageAlign.html
         * @see http://help.adobe.com/en_US/FlashPlatform/reference/actionscript/3/flash/display/Stage.html#align
         * Empty value defaults to "centered" option
         */
        StageAlign: '',

        cameraId: 'AS3webcamObject',
        callTarget: 'webcam',
        bgcolor: '#000000',
        isSwfReady: false,
        isCameraEnabled: false,
        swffile: "sAS3Cam.swf",
        cameraEnabled:   function () { },
        cameraDisabled:  function () { },
        noCameraFound:   function () { },
        isClientReady:   function () { return true; },
        cameraReady:     function () { },
        cameraConnected: function () {
            this.isSwfReady = true;
            var cam = document.getElementById(this.cameraId);

            this.save          = function()  { try { return cam.save();          } catch(e) { this.swfApiFail(e); } }
            this.setCamera     = function(i) { try { return cam.setCamera(i);    } catch(e) { this.swfApiFail(e); } }
            this.getCameraList = function()  { try { return cam.getCameraList(); } catch(e) { this.swfApiFail(e); } }
            this.getResolution = function()  { try { return cam.getResolution(); } catch(e) { this.swfApiFail(e); } },

            this.cameraReady();
        },
        init: function(container, options) {
            if (typeof options === "object") {
                for (var ndx in webcam) {
                    if (options[ndx] !== undefined) {
                        webcam[ndx] = options[ndx];
                    }
                }
            }
            var source = '<object id="'+this.cameraId+'" type="application/x-shockwave-flash" data="'+webcam.swffile+'" width="'+webcam.previewWidth+'" height="'+webcam.previewHeight+'"><param name="movie" value="'+webcam.swffile+'" /><param name="FlashVars" value="callTarget='+this.callTarget+'&resolutionWidth='+webcam.resolutionWidth+'&resolutionHeight='+webcam.resolutionHeight+'&smoothing='+webcam.videoSmoothing+'&deblocking='+webcam.videoDeblocking+'&StageScaleMode='+webcam.StageScaleMode+'&StageAlign='+webcam.StageAlign+'" /><param name="allowScriptAccess" value="always" /><param name="bgcolor" value="'+webcam.bgcolor+'" /></object>';
            $(container).html(source);
            return this;
        },
        swfApiFail: function (e) { },
    	debug:	    function ()  { }
    };
    window.webcam = webcam;
    $.fn.webcam = function(options) { return webcam.init(this, options); };
});
} catch (e) {}
