(function(e){function t(t){for(var n,l,r=t[0],o=t[1],c=t[2],u=0,h=[];u<r.length;u++)l=r[u],Object.prototype.hasOwnProperty.call(s,l)&&s[l]&&h.push(s[l][0]),s[l]=0;for(n in o)Object.prototype.hasOwnProperty.call(o,n)&&(e[n]=o[n]);d&&d(t);while(h.length)h.shift()();return i.push.apply(i,c||[]),a()}function a(){for(var e,t=0;t<i.length;t++){for(var a=i[t],n=!0,r=1;r<a.length;r++){var o=a[r];0!==s[o]&&(n=!1)}n&&(i.splice(t--,1),e=l(l.s=a[0]))}return e}var n={},s={app:0},i=[];function l(t){if(n[t])return n[t].exports;var a=n[t]={i:t,l:!1,exports:{}};return e[t].call(a.exports,a,a.exports,l),a.l=!0,a.exports}l.m=e,l.c=n,l.d=function(e,t,a){l.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:a})},l.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},l.t=function(e,t){if(1&t&&(e=l(e)),8&t)return e;if(4&t&&"object"===typeof e&&e&&e.__esModule)return e;var a=Object.create(null);if(l.r(a),Object.defineProperty(a,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var n in e)l.d(a,n,function(t){return e[t]}.bind(null,n));return a},l.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return l.d(t,"a",t),t},l.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},l.p="/";var r=window["webpackJsonp"]=window["webpackJsonp"]||[],o=r.push.bind(r);r.push=t,r=r.slice();for(var c=0;c<r.length;c++)t(r[c]);var d=o;i.push([0,"chunk-vendors"]),a()})({0:function(e,t,a){e.exports=a("56d7")},"56d7":function(e,t,a){"use strict";a.r(t);a("e260"),a("e6cf"),a("cca6"),a("a79d");var n,s=a("2b0e"),i=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("v-app",[a("v-app-bar",{attrs:{app:"",color:"primary",dark:"",height:"80px"}},[a("div",{staticClass:"d-none d-sm-flex align-center"},[a("v-btn",{attrs:{icon:"",href:"/"}},[a("v-img",{attrs:{alt:"Samar logo",contain:"",src:"sign.png",width:"40"}})],1),a("v-flex",{staticClass:"font-weight-thin headline ml-1 mr-5"},[e._v("SIDCLOUD")])],1),a("v-spacer",{staticClass:"hidden-sm-and-down"}),a("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("play_top",e.last_index)}}},[a("v-icon",[e._v(e._s(e.play_icon))])],1),a("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("jmp",e.last_index-1)}}},[a("v-icon",[e._v("skip_previous")])],1),a("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("jmp",e.last_index+1)}}},[a("v-icon",[e._v("skip_next")])],1),a("v-btn",{staticClass:"mr-2",attrs:{fab:""},on:{click:function(t){return e.click("replay",e.last_index)}}},[a("v-icon",[e._v("replay")])],1),a("v-btn",{staticClass:"hidden-sm-and-down",attrs:{fab:""},on:{click:function(t){return e.click("stop",e.last_index)}}},[a("v-icon",[e._v("stop")])],1),a("v-spacer"),a("v-switch",{staticClass:"hidden-xs-and-down",staticStyle:{"margin-top":"22px"},attrs:{label:e.themeLabel},model:{value:e.darkTheme,callback:function(t){e.darkTheme=t},expression:"darkTheme"}})],1),a("v-main",[a("v-container",{attrs:{fluid:""}},[a("v-row",e._l(e.releases,(function(t,n){return a("v-col",{key:t.ReleaseID},[a("v-row",{staticClass:"my-1",attrs:{align:"center",justify:"center"}},[a("v-hover",{scopedSlots:e._u([{key:"default",fn:function(s){var i=s.hover;return[a("v-card",{staticClass:"card-outter",attrs:{elevation:i?5:2,disabled:e.cardDisabled(n),id:"cnr"+t.ReleaseID.toString(),"min-height":"450",width:"300"}},[a("v-img",{style:e.cursorOverImg(n),attrs:{elevation:i?20:2,src:t.ReleaseScreenShot,width:"300",height:"212"},on:{click:function(t){return e.click("play",n)}}},[i?a("v-overlay",{attrs:{absolute:"",color:"#222222"}},[a("v-btn",{staticClass:"mr-2",attrs:{fab:""}},[e.playingNow[n]?a("v-icon",[e._v("pause")]):e._e(),e.playingNow[n]?e._e():a("v-icon",[e._v("play_arrow")])],1)],1):e._e()],1),a("v-progress-linear",{attrs:{value:e.current_time(n),indeterminate:e.indeterminate(n),active:n==e.last_index,height:"10"}}),a("v-rating",{staticClass:"d-flex justify-center mt-1",attrs:{value:t.Rating,size:"24",length:"10",readonly:"",dense:"","half-increments":""}}),a("v-card-text",{staticClass:"mx-3 mt-1 pa-0 ma-0 caption",domProps:{textContent:e._s(e.releaseDate(n))}}),a("v-card-title",{staticClass:"mx-3 mt-1 pa-0 ma-0",domProps:{textContent:e._s(e.releaseName(n))}}),a("v-card-subtitle",{staticClass:"mx-3 mb-2 pa-0 ma-0"},[e._v(" "+e._s(e.releasedWithComma(n))+" ")]),a("v-card-text",{staticClass:"mx-3 mb-2 pa-0 ma-0 font-italic font-weight-medium"},[e._v(" Music by "+e._s(e.creditsWithComma(n))+" ")]),a("v-card-text",{staticClass:"mx-3 pa-0 ma-0 caption",domProps:{textContent:e._s(t.ReleasedAt.substring(0,40))}}),a("v-card-actions",{staticClass:"card-actions"},[a("v-btn",{attrs:{text:"",color:"deep-purple accent-4",href:e.linkToSidcloudId(t.ReleaseID),link:"",target:"_blank"}},[e._v(" Link "),a("v-icon",{attrs:{small:""}},[e._v("launch")])],1),a("v-btn",{attrs:{text:"",color:"deep-purple accent-4",href:e.linkToCsdbId(t.ReleaseID),link:"",target:"_blank"}},[e._v(" CSDB "),a("v-icon",{attrs:{small:""}},[e._v("launch")])],1)],1)],1)]}}],null,!0)})],1)],1)})),1)],1)],1),a("v-bottom-navigation",{attrs:{app:"",grow:""}},[a("v-container",{staticStyle:{margin:"0px",padding:"0px",width:"100%"},attrs:{fluid:""}},[a("v-layout",{attrs:{wrap:"","justify-center":""}},[a("v-progress-linear",{attrs:{value:e.timeCurrent/e.timeDuration*100,"buffer-value":e.timeBuffered,stream:e.music_play,height:"10",indeterminate:e.music_loading},on:{click:function(t){return e.sliderClick(t)}}}),a("v-btn",{staticClass:"mt-1 font-weight-thin title",attrs:{text:"",block:"",link:"",target:"_blank"},on:{click:function(t){e.$vuetify.goTo("#cnr"+e.releases[e.last_index].ReleaseID.toString(),e.gotoOptions)}}},[e._v(" "+e._s(e.title_playing)+" ")])],1)],1)],1),a("audio",{attrs:{id:"radio",preload:"none"}},[a("source",{attrs:{src:e.audio_url,type:"audio/wav"}})])],1)},l=[],r=(a("cb29"),a("b0c0"),a("d3b7"),a("ac1f"),a("25f0"),a("4d90"),a("5319"),a("bc3a")),o=a.n(r),c={name:"App",data:function(){return{releases:null,playingNow:[],title_playing:"SIDCLOUD - YOUR SID NEWSPAPER",audio_url:"",music_play:!1,paused:!1,player_type:"sidplayfp",last_index:0,music_ended:!1,music_loading:!1,timeDuration:300,timeCurrent:0,timeBuffered:0,newTimeCurrent:0,linkToCsdb:"https://csdb.dk/",playedOnce:!1,darkTheme:!1,path_id:0}},watch:{darkTheme:function(e){localStorage.setItem("darkTheme",e),this.$vuetify.theme.dark=e}},computed:{gotoOptions:function(){return{duration:300,offset:20,easing:"easeInOutCubic"}},play_icon:function(){return this.music_play&&!this.paused?"pause":"play_arrow"},themeLabel:function(){var e="Dark";return"xs"!=this.$vuetify.breakpoint.name&&"sm"!=this.$vuetify.breakpoint.name||(e=""),e}},methods:{sliderClick:function(){},toggleTheme:function(){this.themeDark?(this.$vuetify.theme.dark=!1,this.themeDark=!1):(this.$vuetify.theme.dark=!0,this.themeDark=!0),localStorage.setItem("themeDark",this.themeDark)},releaseName:function(e){return this.releases[e].ReleaseName.length<25?this.releases[e].ReleaseName:this.releases[e].ReleaseName.substring(0,25)+"..."},releaseDate:function(e){var t=this.releases[e].ReleaseYear.toString(),a=this.releases[e].ReleaseMonth.toString();a=a.padStart(2,"0");var n=this.releases[e].ReleaseDay.toString();return n=n.padStart(2,"0"),0==this.releases[e].ReleaseMonth&&(a="??"),0==this.releases[e].ReleaseDay&&(n="??"),0==this.releases[e].ReleaseYear&&(t="????"),this.releases[e].ReleaseType+" / "+t+"-"+a+"-"+n},playTimeChange:function(){this.timeCurrent=n.currentTime,n.seekable.length>0&&(this.timeBuffered=n.seekable.end(n.seekable.length-1))},clearPlayingNow:function(){for(var e=0;e<this.playingNow.length;e++)this.playingNow[e]=!1},setPlayingNow:function(e){this.clearPlayingNow(),this.playingNow[e]=!0},linkToCsdbId:function(e){var t="";return t="https://csdb.dk/release/?id="+e,t},linkToSidcloudId:function(e){var t="";return t="https://sidcloud.net/?id="+e,t},cardDisabled:function(e){return!this.releases[e].WAVCached},cursorOverImg:function(e){return this.releases[e].WAVCached?"cursor: pointer":""},current_time:function(e){return e==this.last_index?(this.timeCurrent/this.timeDuration*100).toString():"0"},indeterminate:function(e){return!(e!=this.last_index||!this.music_loading)},releasedWithComma:function(e){var t="";if(null!=this.releases[e].ReleasedBy)for(var a=0;a<this.releases[e].ReleasedBy.length;a++)t+=0!=a?", "+this.releases[e].ReleasedBy[a]:this.releases[e].ReleasedBy[a];return t},creditsWithComma:function(e){var t=" ";if(null!=this.releases[e].Credits)for(var a=0;a<this.releases[e].Credits.length;a++)t+=0!=a?", "+this.releases[e].Credits[a]:this.releases[e].Credits[a];return t},AudioUrl:function(){null==o.a.defaults.baseURL||void 0===o.a.defaults.baseURL?this.audio_url="/api/v1/audio/"+this.releases[this.last_index].ReleaseID:this.audio_url=o.a.defaults.baseURL+"/api/v1/audio/"+this.releases[this.last_index].ReleaseID},ended:function(){console.log("player event: ended"),this.clearPlayingNow(),this.music_ended=!0,this.click("jmp",this.last_index+1)},canplay:function(){console.log("player event: canplay")},timeupdate:function(){console.log("player event: timeupdate"),this.timeCurrent=n.currentTime},playing:function(){console.log("player event: playing"),this.setPlayingNow(this.last_index),this.paused=!1,this.music_play=!0,this.playedOnce=!0,this.music_loading=!1},canplaythrough:function(){console.log("player event: canplaythrough")},play:function(){console.log("player event: play")},pause:function(){console.log("player event: pause")},loadedmetadata:function(){console.log("player event: loadedmetadata")},loadeddata:function(){console.log("player event: loadeddata")},waiting:function(){console.log("player event: waiting")},audioprocess:function(){console.log("player event: audioprocess")},complete:function(){console.log("player event: complete")},emptied:function(){console.log("player event: emptied")},ratechange:function(){console.log("player event: ratechange")},seeked:function(){console.log("player event: seeked")},seeking:function(){console.log("player event: seeking")},stalled:function(){console.log("player event: stalled")},suspend:function(){console.log("player event: suspend")},volumechange:function(){console.log("player event: volumechange")},durationchange:function(){console.log("player event: durationchange"),console.log("durationchange: readed "+n.duration),n.duration>300?(this.timeDuration=300,console.log("durationchange: changed to "+this.timeDuration)):this.timeDuration=n.duration},keydown:function(e){switch(e.code){case"MediaPlayPause":console.log("window event: keydown"),console.log(e.code),this.click("play",this.last_index);break;case"MediaTrackNext":this.click("jmp",this.last_index+1);break;case"MediaTrackPrevious":console.log("window event: keydown"),console.log(e.code),this.click("jmp",this.last_index-1);break}},click:function(e,t){switch(e){case"stop":console.log("job: stop"),this.clearPlayingNow(),this.audio_url="",n.load(),this.paused=!1,this.music_play=!1,this.music_ended=!0;break;case"jmp":case"replay":if(console.log("job: jmp"),t==this.last_index&&"jmp"==e&&this.playedOnce&&!this.music_loading||t<0||t>=this.releases.length)return;if(t>this.last_index)while(!this.releases[t].WAVCached){if(!(t<80))return;t++}else if(t<this.last_index)while(!this.releases[t].WAVCached){if(!(t>0))return;t--}this.clearPlayingNow(),n.pause(),this.timeCurrent=0,n.currentTime=0,this.paused=!1,this.music_play=!1,("xs"==this.$vuetify.breakpoint.name||"sm"==this.$vuetify.breakpoint.name)&&this.releases[t].ReleaseName.length>27?this.title_playing=this.releases[t].ReleaseName.substring(0,27)+" ...":this.title_playing=this.releases[t].ReleaseName,this.last_index=t,this.music_loading=!0,this.linkToCsdb="https://csdb.dk/release/?id="+this.releases[this.last_index].ReleaseID,this.AudioUrl(),n.load(),console.log("Loading..."),n.play(),this.music_ended=!1;break;case"play_top":case"play":if(console.log("job: play"),this.paused&&t==this.last_index)this.paused=!1,this.music_ended=!1,n.play();else if(this.music_play&&t==this.last_index)this.clearPlayingNow(),n.pause(),this.paused=!0,this.music_play=!1;else{if(this.clearPlayingNow(),n.pause(),this.timeCurrent=0,n.currentTime=0,this.paused=!1,this.music_play=!1,this.path_id>0&&"play_top"==e){for(var a=0;a<this.releases.length;a++)if(this.releases[a].ReleaseID==this.path_id){this.last_index=a,t=a,this.path_id=0,this.$vuetify.goTo("#cnr"+this.releases[this.last_index].ReleaseID.toString(),this.gotoOptions);break}}else this.last_index=t;("xs"==this.$vuetify.breakpoint.name||"sm"==this.$vuetify.breakpoint.name)&&this.releases[t].ReleaseName.length>27?this.title_playing=this.releases[t].ReleaseName.substring(0,27)+" ...":this.title_playing=this.releases[t].ReleaseName,this.AudioUrl(),this.music_loading=!0,this.linkToCsdb="https://csdb.dk/release/?id="+this.releases[this.last_index].ReleaseID,n.load(),console.log("Loading..."),n.play(),this.music_ended=!1}break}}},created:function(){var e=this,t=this.$route.query.id;null!=t?t>0?(console.log("Params: ",t),o.a.get("/api/v1/csdb_release/"+t).then((function(t){console.log("Response: "),console.log(t.data),e.releases=t.data,e.playingNow=new Array(e.releases.length).fill(!1);var a=window.location.pathname;a=a.replace(/\D/g,""),e.path_id=parseInt(a,10)})).catch((function(e){console.log(e)}))):o.a.get("/api/v1/csdb_releases").then((function(t){console.log("Response: "),console.log(t.data),e.releases=t.data,e.playingNow=new Array(e.releases.length).fill(!1);var a=window.location.pathname;a=a.replace(/\D/g,""),e.path_id=parseInt(a,10)})).catch((function(e){console.log(e)})):o.a.get("/api/v1/csdb_releases").then((function(t){console.log("Response: "),console.log(t.data),e.releases=t.data,e.playingNow=new Array(e.releases.length).fill(!1);var a=window.location.pathname;a=a.replace(/\D/g,""),e.path_id=parseInt(a,10)})).catch((function(e){console.log(e)}))},mounted:function(){this.darkTheme="true"==localStorage.getItem("darkTheme"),n=document.getElementById("radio"),n.addEventListener("ended",this.ended),n.addEventListener("canplay",this.canplay),n.addEventListener("playing",this.playing),n.addEventListener("canplaythrough",this.canplaythrough),n.addEventListener("play",this.play),n.addEventListener("pause",this.pause),n.addEventListener("loadedmetadata",this.loadedmetadata),n.addEventListener("loadeddata",this.loadeddata),n.addEventListener("waiting",this.waiting),n.addEventListener("audioprocess",this.audioprocess),n.addEventListener("complete",this.complete),n.addEventListener("emptied",this.emptied),n.addEventListener("ratechange",this.ratechange),n.addEventListener("seeked",this.seeked),n.addEventListener("seeking",this.seeking),n.addEventListener("stalled",this.stalled),n.addEventListener("volumechange",this.volumechange),n.addEventListener("durationchange",this.durationchange),window.addEventListener("keydown",this.keydown),setInterval(this.playTimeChange,1e3)}},d=c,u=(a("b777"),a("2877")),h=a("6544"),p=a.n(h),m=a("7496"),g=a("40dc"),v=a("b81c"),f=a("8336"),y=a("b0af"),_=a("99d9"),b=a("62ad"),k=a("a523"),w=a("0e8f"),x=a("ce87"),C=a("132d"),R=a("adda"),D=a("a722"),S=a("f6c4"),T=a("a797"),N=a("8e36"),I=a("1d4d"),L=a("0fd9"),V=a("2fa4"),O=a("b73d"),j=Object(u["a"])(d,i,l,!1,null,"7ccb0801",null),P=j.exports;p()(j,{VApp:m["a"],VAppBar:g["a"],VBottomNavigation:v["a"],VBtn:f["a"],VCard:y["a"],VCardActions:_["a"],VCardSubtitle:_["b"],VCardText:_["c"],VCardTitle:_["d"],VCol:b["a"],VContainer:k["a"],VFlex:w["a"],VHover:x["a"],VIcon:C["a"],VImg:R["a"],VLayout:D["a"],VMain:S["a"],VOverlay:T["a"],VProgressLinear:N["a"],VRating:I["a"],VRow:L["a"],VSpacer:V["a"],VSwitch:O["a"]});var E=a("8c4f"),A=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"main"},[a("Main")],1)},$=[],M=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("v-container")},B=[],U={name:"Main",data:function(){return{}}},W=U,H=Object(u["a"])(W,M,B,!1,null,null,null),Y=H.exports;p()(H,{VContainer:k["a"]});var J={name:"Home",components:{Main:Y}},q=J,z=Object(u["a"])(q,A,$,!1,null,null,null),F=z.exports;s["a"].use(E["a"]);var G=[{path:"/:id",name:"Home",component:F}],K=new E["a"]({mode:"history",base:"/",routes:G}),Q=K,X=a("2f62");s["a"].use(X["a"]);var Z=new X["a"].Store({state:{},mutations:{},actions:{},modules:{}}),ee=a("f309");a("d1e78");s["a"].use(ee["a"]);var te=new ee["a"]({icons:{iconfont:"md"}});s["a"].config.productionTip=!1,new s["a"]({router:Q,store:Z,vuetify:te,axios:o.a,render:function(e){return e(P)}}).$mount("#app")},b231:function(e,t,a){},b777:function(e,t,a){"use strict";a("b231")}});
//# sourceMappingURL=app.c507bb7a.js.map