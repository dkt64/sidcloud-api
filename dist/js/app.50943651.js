(function(e){function t(t){for(var a,o,r=t[0],l=t[1],c=t[2],d=0,p=[];d<r.length;d++)o=r[d],Object.prototype.hasOwnProperty.call(i,o)&&i[o]&&p.push(i[o][0]),i[o]=0;for(a in l)Object.prototype.hasOwnProperty.call(l,a)&&(e[a]=l[a]);u&&u(t);while(p.length)p.shift()();return s.push.apply(s,c||[]),n()}function n(){for(var e,t=0;t<s.length;t++){for(var n=s[t],a=!0,o=1;o<n.length;o++){var l=n[o];0!==i[l]&&(a=!1)}a&&(s.splice(t--,1),e=r(r.s=n[0]))}return e}var a={},i={app:0},s=[];function o(e){return r.p+"js/"+({about:"about"}[e]||e)+"."+{about:"e154e38b"}[e]+".js"}function r(t){if(a[t])return a[t].exports;var n=a[t]={i:t,l:!1,exports:{}};return e[t].call(n.exports,n,n.exports,r),n.l=!0,n.exports}r.e=function(e){var t=[],n=i[e];if(0!==n)if(n)t.push(n[2]);else{var a=new Promise((function(t,a){n=i[e]=[t,a]}));t.push(n[2]=a);var s,l=document.createElement("script");l.charset="utf-8",l.timeout=120,r.nc&&l.setAttribute("nonce",r.nc),l.src=o(e);var c=new Error;s=function(t){l.onerror=l.onload=null,clearTimeout(d);var n=i[e];if(0!==n){if(n){var a=t&&("load"===t.type?"missing":t.type),s=t&&t.target&&t.target.src;c.message="Loading chunk "+e+" failed.\n("+a+": "+s+")",c.name="ChunkLoadError",c.type=a,c.request=s,n[1](c)}i[e]=void 0}};var d=setTimeout((function(){s({type:"timeout",target:l})}),12e4);l.onerror=l.onload=s,document.head.appendChild(l)}return Promise.all(t)},r.m=e,r.c=a,r.d=function(e,t,n){r.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:n})},r.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},r.t=function(e,t){if(1&t&&(e=r(e)),8&t)return e;if(4&t&&"object"===typeof e&&e&&e.__esModule)return e;var n=Object.create(null);if(r.r(n),Object.defineProperty(n,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var a in e)r.d(n,a,function(t){return e[t]}.bind(null,a));return n},r.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return r.d(t,"a",t),t},r.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},r.p="/",r.oe=function(e){throw console.error(e),e};var l=window["webpackJsonp"]=window["webpackJsonp"]||[],c=l.push.bind(l);l.push=t,l=l.slice();for(var d=0;d<l.length;d++)t(l[d]);var u=c;s.push([0,"chunk-vendors"]),n()})({0:function(e,t,n){e.exports=n("56d7")},"56d7":function(e,t,n){"use strict";n.r(t);n("e260"),n("e6cf"),n("cca6"),n("a79d");var a=n("2b0e"),i=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("v-app",[n("v-app-bar",{attrs:{app:"",color:"primary",dark:"",height:"80px"}},[n("div",{staticClass:"d-none d-sm-flex align-center"},[n("v-img",{staticClass:"mr-1",attrs:{alt:"Samar logo",contain:"",src:"sign.png",width:"40"}}),n("v-flex",{staticClass:"font-weight-thin headline mr-5"},[e._v("SIDCLOUD")])],1),n("v-spacer"),n("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("play",e.last_index)}}},[n("v-icon",[e._v(e._s(e.play_icon))])],1),n("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("jmp",e.last_index-1)}}},[n("v-icon",[e._v("skip_previous")])],1),n("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("jmp",e.last_index+1)}}},[n("v-icon",[e._v("skip_next")])],1),n("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("replay",e.last_index)}}},[n("v-icon",[e._v("replay")])],1),n("v-btn",{attrs:{fab:""},on:{click:function(t){return e.click("stop",e.last_index)}}},[n("v-icon",[e._v("stop")])],1),n("v-spacer"),n("v-switch",{staticClass:"ml-2 hidden-sm-and-down",staticStyle:{"margin-top":"22px"},attrs:{label:"Theme"},model:{value:e.$vuetify.theme.dark,callback:function(t){e.$set(e.$vuetify.theme,"dark",t)},expression:"$vuetify.theme.dark"}})],1),n("v-content",[n("v-container",{attrs:{fluid:""}},[n("v-row",{attrs:{dense:""}},e._l(e.releases,(function(t,a){return n("v-col",{key:t.ReleaseID},[n("v-card",{staticClass:"mx-auto mb-5",attrs:{"min-height":"420",width:"320"},on:{click:function(t){return e.click("jmp",a)}}},[n("v-img",{attrs:{src:t.ReleaseScreenShot,width:"320",height:"227"}}),n("v-progress-linear",{attrs:{value:e.current_time(a),indeterminate:e.indeterminate(a),height:"8"}}),n("v-rating",{staticClass:"d-flex justify-center mt-1",attrs:{value:t.Rating,size:"27",length:"10",readonly:"",dense:""}}),n("v-card-title",{domProps:{textContent:e._s(t.ReleaseName.substring(0,27))}}),n("v-card-subtitle",e._l(t.ReleasedBy,(function(t,a){return n("v-div",{key:t},[e._v(e._s(e.nameWithComma(a))+e._s(t))])})),1),n("v-card-text",e._l(t.Credits,(function(t,a){return n("v-div",{key:t,staticClass:"font-italic font-weight-medium"},[e._v(e._s(e.nameWithComma(a))+e._s(t))])})),1)],1)],1)})),1)],1)],1),n("v-bottom-navigation",{attrs:{app:"",grow:""}},[n("v-container",{staticStyle:{margin:"0px",padding:"0px",width:"100%"},attrs:{fluid:""}},[n("v-layout",{attrs:{wrap:"","justify-center":""}},[n("v-progress-linear",{attrs:{value:e.timeCurrent/e.timeDuration*100,height:"8",indeterminate:e.music_loading}}),n("v-btn",{staticClass:"mt-1 font-weight-thin title",attrs:{href:e.linkToCsdb,link:"",text:"",target:"_blank"}},[e._v(" "+e._s(e.title_playing)+" ")])],1)],1)],1),n("audio",{attrs:{id:"radio",preload:"none"}},[n("source",{attrs:{src:e.audio_url,type:"audio/wav"}})])],1)},s=[],o=(n("d3b7"),n("25f0"),n("bc3a")),r=n.n(o),l=document.getElementById("radio"),c={name:"App",data:function(){return{releases:null,title_playing:"SIDCLOUD - YOUR SID NEWSPAPER",audio_url:"",music_play:!1,paused:!1,player_type:"sidplayfp",last_index:0,music_ended:!1,music_loading:!1,timeDuration:305,timeCurrent:0,linkToCsdb:"https://csdb.dk/",playedOnce:!1}},computed:{play_icon:function(){return this.music_play&&!this.paused?"pause":"play_arrow"}},methods:{current_time:function(e){return this.music_play&&e==this.last_index?(this.timeCurrent/this.timeDuration*100).toString():"0"},indeterminate:function(e){return!(e!=this.last_index||!this.music_loading)},nameWithComma:function(e){return 0==e?"":", "},AudioUrl:function(){null==r.a.defaults.baseURL||void 0===r.a.defaults.baseURL?this.audio_url="/api/v1/audio/"+this.player_type:this.audio_url=r.a.defaults.baseURL+"/api/v1/audio/"+this.player_type},ended:function(){console.log("player event: ended"),this.music_ended=!0,this.click("jmp",this.last_index+1)},canplay:function(){console.log("player event: canplay"),this.music_ended||(l.play(),console.log("player.play()..."),this.paused=!1,this.music_play=!0,this.playedOnce=!0,this.linkToCsdb="https://csdb.dk/release/?id="+this.releases[this.last_index].ReleaseID)},timeupdate:function(){this.timeCurrent=l.currentTime},playing:function(){console.log("player event: playing"),this.music_loading=!1},canplaythrough:function(){console.log("player event: canplaythrough")},play:function(){console.log("player event: play")},pause:function(){console.log("player event: pause")},loadedmetadata:function(){console.log("player event: loadedmetadata")},loadeddata:function(){console.log("player event: loadeddata")},waiting:function(){console.log("player event: waiting")},audioprocess:function(){console.log("player event: audioprocess")},complete:function(){console.log("player event: complete")},emptied:function(){console.log("player event: emptied")},ratechange:function(){console.log("player event: ratechange")},seeked:function(){console.log("player event: seeked")},seeking:function(){console.log("player event: seeking")},stalled:function(){console.log("player event: stalled")},suspend:function(){console.log("player event: suspend")},volumechange:function(){console.log("player event: volumechange")},click:function(e,t){var n,a=this;switch(e){case"stop":l.pause(),l.currentTime=0,this.paused=!1,this.music_play=!1,this.music_ended=!0;break;case"jmp":case"replay":if(t==this.last_index&&"jmp"==e&&this.playedOnce&&!this.music_loading||t<0||t>=this.releases.length)return;l.pause(),l.currentTime=0,this.paused=!1,this.music_play=!1,this.title_playing=this.releases[t].ReleaseName,this.last_index=t,n="/api/v1/audio?sid_url="+this.releases[t].DownloadLinks[0],this.music_loading=!0,r.a.post(n).then((function(e){console.log(e.data),a.AudioUrl(),l.load(),console.log("Loading..."),a.music_ended=!1}));break;case"play":this.paused?(this.paused=!1,this.music_ended=!1,this.AudioUrl(),l.play()):this.music_play?(l.pause(),this.paused=!0,this.musicPlay=!1):(l.pause(),l.currentTime=0,this.paused=!1,this.music_play=!1,this.title_playing=this.releases[t].ReleaseName,this.last_index=t,n="/api/v1/audio?sid_url="+this.releases[t].DownloadLinks[0],this.music_loading=!0,r.a.post(n).then((function(e){console.log(e.data),a.AudioUrl(),l.load(),console.log("Loading..."),a.music_ended=!1})));break}}},created:function(){var e=this;r.a.get("/api/v1/csdb_releases").then((function(t){console.log("Response: "),console.log(t.data),e.releases=t.data})).catch((function(e){console.log(e)})),this.AudioUrl()},mounted:function(){l=document.getElementById("radio"),l.addEventListener("ended",this.ended),l.addEventListener("canplay",this.canplay),l.addEventListener("timeupdate",this.timeupdate),l.addEventListener("playing",this.playing),l.addEventListener("canplaythrough",this.canplaythrough),l.addEventListener("play",this.play),l.addEventListener("pause",this.pause),l.addEventListener("loadedmetadata",this.loadedmetadata),l.addEventListener("loadeddata",this.loadeddata),l.addEventListener("waiting",this.waiting),l.addEventListener("audioprocess",this.audioprocess),l.addEventListener("complete",this.complete),l.addEventListener("emptied",this.emptied),l.addEventListener("ratechange",this.ratechange),l.addEventListener("seeked",this.seeked),l.addEventListener("seeking",this.seeking),l.addEventListener("stalled",this.stalled),l.addEventListener("suspend",this.suspend),l.addEventListener("volumechange",this.volumechange)}},d=c,u=n("2877"),p=n("6544"),h=n.n(p),m=n("7496"),v=n("40dc"),f=n("b81c"),g=n("8336"),y=n("b0af"),_=n("99d9"),b=n("62ad"),k=n("a523"),w=n("a75b"),x=n("0e8f"),C=n("132d"),L=n("adda"),E=n("a722"),j=n("8e36"),V=n("1d4d"),O=n("0fd9"),S=n("2fa4"),R=n("b73d"),T=Object(u["a"])(d,i,s,!1,null,null,null),P=T.exports;h()(T,{VApp:m["a"],VAppBar:v["a"],VBottomNavigation:f["a"],VBtn:g["a"],VCard:y["a"],VCardSubtitle:_["a"],VCardText:_["b"],VCardTitle:_["c"],VCol:b["a"],VContainer:k["a"],VContent:w["a"],VFlex:x["a"],VIcon:C["a"],VImg:L["a"],VLayout:E["a"],VProgressLinear:j["a"],VRating:V["a"],VRow:O["a"],VSpacer:S["a"],VSwitch:R["a"]});var A=n("8c4f"),D=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"home"},[a("img",{attrs:{alt:"Vue logo",src:n("cf05")}}),a("HelloWorld",{attrs:{msg:"Welcome to Your Vue.js App"}})],1)},U=[],I=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("v-container")},W=[],$={name:"HelloWorld",data:function(){return{}}},B=$,H=Object(u["a"])(B,I,W,!1,null,null,null),N=H.exports;h()(H,{VContainer:k["a"]});var M={name:"Home",components:{HelloWorld:N}},J=M,Y=Object(u["a"])(J,D,U,!1,null,null,null),q=Y.exports;a["a"].use(A["a"]);var z=[{path:"/",name:"Home",component:q},{path:"/about",name:"About",component:function(){return n.e("about").then(n.bind(null,"f820"))}}],F=new A["a"]({mode:"history",base:"/",routes:z}),G=F,K=n("2f62");a["a"].use(K["a"]);var Q=new K["a"].Store({state:{},mutations:{},actions:{},modules:{}}),X=n("f309");n("d1e78");a["a"].use(X["a"]);var Z=new X["a"]({icons:{iconfont:"md"}});a["a"].config.productionTip=!1,new a["a"]({router:G,store:Q,vuetify:Z,axios:r.a,render:function(e){return e(P)}}).$mount("#app")},cf05:function(e,t,n){e.exports=n.p+"img/logo.82b9c7a5.png"}});
//# sourceMappingURL=app.50943651.js.map