(function(e){function a(a){for(var t,s,l=a[0],d=a[1],i=a[2],u=0,p=[];u<l.length;u++)s=l[u],Object.prototype.hasOwnProperty.call(r,s)&&r[s]&&p.push(r[s][0]),r[s]=0;for(t in d)Object.prototype.hasOwnProperty.call(d,t)&&(e[t]=d[t]);c&&c(a);while(p.length)p.shift()();return o.push.apply(o,i||[]),n()}function n(){for(var e,a=0;a<o.length;a++){for(var n=o[a],t=!0,l=1;l<n.length;l++){var d=n[l];0!==r[d]&&(t=!1)}t&&(o.splice(a--,1),e=s(s.s=n[0]))}return e}var t={},r={app:0},o=[];function s(a){if(t[a])return t[a].exports;var n=t[a]={i:a,l:!1,exports:{}};return e[a].call(n.exports,n,n.exports,s),n.l=!0,n.exports}s.m=e,s.c=t,s.d=function(e,a,n){s.o(e,a)||Object.defineProperty(e,a,{enumerable:!0,get:n})},s.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},s.t=function(e,a){if(1&a&&(e=s(e)),8&a)return e;if(4&a&&"object"===typeof e&&e&&e.__esModule)return e;var n=Object.create(null);if(s.r(n),Object.defineProperty(n,"default",{enumerable:!0,value:e}),2&a&&"string"!=typeof e)for(var t in e)s.d(n,t,function(a){return e[a]}.bind(null,t));return n},s.n=function(e){var a=e&&e.__esModule?function(){return e["default"]}:function(){return e};return s.d(a,"a",a),a},s.o=function(e,a){return Object.prototype.hasOwnProperty.call(e,a)},s.p="/";var l=window["webpackJsonp"]=window["webpackJsonp"]||[],d=l.push.bind(l);l.push=a,l=l.slice();for(var i=0;i<l.length;i++)a(l[i]);var c=d;o.push([0,"chunk-vendors"]),n()})({0:function(e,a,n){e.exports=n("56d7")},"034f":function(e,a,n){"use strict";var t=n("64a9"),r=n.n(t);r.a},"56d7":function(e,a,n){"use strict";n.r(a);n("cadf"),n("551c"),n("f751"),n("097d");var t=n("2b0e"),r=function(){var e=this,a=e.$createElement,n=e._self._c||a;return n("div",{attrs:{id:"app"},on:{drop:function(a){return a.preventDefault(),e.addFile(a)},dragover:function(e){e.preventDefault()},"!dragover":function(a){return e.dragingOn(a)},"!drop":function(a){return e.dragingOff(a)}}},[n("h1",[e._v("Welcome to SIDCloud")]),n("img",{attrs:{src:e.csdb_release_screenshot}}),n("p"),n("p",[e._v(e._s(e.csdb_release_name)+" "+e._s(e.csdb_release_group))]),n("p"),n("audio",{attrs:{type:"audio/wav",src:e.audio_src,id:"radio",controls:"",preload:"none",loop:""}},[e._v("Audio element is not supported on Your browser :(")]),n("p"),n("b-container",{staticClass:"bv-example-row"},[n("b-row",{staticClass:"my-1"},[n("b-col",{attrs:{sm:"3"}},[n("button",{staticClass:"btn btn-success",staticStyle:{"margin-right":"10px","margin-left":"10px"},attrs:{type:"button",disabled:e.dragin},on:{click:e.Link}},[e._v("Load and stream")])]),n("b-col",{attrs:{sm:"3"}},[n("button",{staticClass:"btn btn-success",staticStyle:{"margin-right":"20px"},attrs:{type:"button",disabled:e.dragin},on:{click:e.Next}},[e._v("Goto next release")])])],1),n("p"),n("b-col",{attrs:{sm:"3"}},[n("input",{directives:[{name:"model",rawName:"v-model",value:e.sid_link,expression:"sid_link"}],staticClass:"form-control",staticStyle:{},attrs:{placeholder:"Paste SID/PRG link and press Enter to play or Drag your SID/PRG file here",disabled:e.dragin},domProps:{value:e.sid_link},on:{keyup:function(a){return!a.type.indexOf("key")&&e._k(a.keyCode,"enter",13,a.key,"Enter")?null:e.Link(a)},input:function(a){a.target.composing||(e.sid_link=a.target.value)}}})])],1),n("div",{attrs:{id:"log"}},[e._v(e._s(e.log))])],1)},o=[],s=(n("4917"),n("a481"),n("bc3a")),l=n.n(s),d={data:function(){return{handle_id:0,csdb_releases:null,csdb_release:null,csdb_release_id:null,csdb_release_data:null,csdb_download_links:null,csdb_release_name:"",csdb_release_group:"",csdb_release_credits:"",csdb_release_screenshot:"",release_nr:1,info:null,sid_link:null,response_from_server:null,query_url:"",file:null,dragin:!1,log:"log",audio_src_org:"http://sidcloud.net/api/v1/audio",audio_src:"http://sidcloud.net/api/v1/audio"}},name:"app",components:{},created:function(){this.GetCSDBData()},methods:{dragingOn:function(){this.dragin=!0},dragingOff:function(){this.dragin=!1},addFile:function(e){var a=this,n="http://sidcloud.net/api/v1/audio",t=document.getElementById("radio");t.pause(),t.currentTime=0;var r=new FormData;r.append("file",e.dataTransfer.files[0]);var o={headers:{"content-type":e.dataTransfer.files[0].type}};l.a.put(n,r,o).then((function(e){console.log("File sent. Response: ",e.data),a.response_from_server=e.data,t.load(),t.play()}))},Next:function(){this.GetCSDBData()},GetCSDBData:function(){var e,a=this;this.csdb_release_name="",this.csdb_release_group="",this.csdb_release_credits="",this.csdb_release_screenshot="",e="http://sidcloud.net/api/v1/csdb_releases",console.log("GetCSDBData() "+e),l.a.get(e).then((function(n){a.csdb_releases=n.data;var t,r=new DOMParser;for(t=0;t<100;t++)if(a.csdb_release=r.parseFromString(a.csdb_releases,"application/xml").getElementsByTagName("description")[a.release_nr++].childNodes[0].nodeValue,a.csdb_release.indexOf(">C64 Demo</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 One-File Demo</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Intro</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 4K Intro</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Crack intro</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 REU Release</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Music</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Music Collection</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Graphics Collection</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Diskmag</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Charts</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Invitation</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 1K Intro</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Fake Demo</a><br /><a href=")>0||a.csdb_release.indexOf(">C128 Release</a><br /><a href=")>0||a.csdb_release.indexOf(">SuperCPU Release</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 DTV</a><br /><a href=")>0||a.csdb_release.indexOf(">C64 Fake Demo</a><br /><a href=")>0)break;a.csdb_release_id=r.parseFromString(a.csdb_releases,"application/xml").getElementsByTagName("link")[a.release_nr].childNodes[0].nodeValue,a.csdb_release_id=parseInt(a.csdb_release_id.replace(/^[^0-9]+/,""),10),e="http://sidcloud.net/api/v1/csdb_release?id="+a.csdb_release_id+"&depth=2",l.a.post(e).then((function(e){a.csdb_release_data=e.data;var n,t,o=(a.csdb_release_data.match(/<Link>/g)||[]).length;console.log("Liczba linków: "+o),a.csdb_download_links=o;var s=!1;for(a.sid_link="",n=0;n<o;n++)if(t=r.parseFromString(a.csdb_release_data,"application/xml").getElementsByTagName("Link")[n].childNodes[0].nodeValue,t.indexOf(".sid")>0){s=!0;break}if(!s)for(n=0;n<o;n++)if(t=r.parseFromString(a.csdb_release_data,"application/xml").getElementsByTagName("Link")[n].childNodes[0].nodeValue,t.indexOf(".prg")>0)break;if(s){a.sid_link=t,a.csdb_release_name=r.parseFromString(a.csdb_release_data,"application/xml").getElementsByTagName("Name")[0].childNodes[0].nodeValue;var l=(a.csdb_release_data.match(/<Group>/g)||[]).length;console.log("Liczba grup: "+l),a.csdb_release_group=l>0?"by ":"";var d=" ";for(n=2;n<l+2;n++)n>2&&(d=", "),a.csdb_release_group+=d+r.parseFromString(a.csdb_release_data,"application/xml").getElementsByTagName("Name")[n].childNodes[0].nodeValue;a.csdb_release_screenshot=r.parseFromString(a.csdb_release_data,"application/xml").getElementsByTagName("ScreenShot")[0].childNodes[0].nodeValue,console.log("Screenshot: "+a.csdb_release_screenshot)}else a.GetCSDBData()}))}))},Link:function(){var e=this,a="http://sidcloud.net/api/v1/audio?sid_url="+this.sid_link;console.log("Query = "+a),document.getElementById("log").innerHTML="["+Date.now()+"] start  ";var n=document.getElementById("radio");n.pause(),n.currentTime=0,n.addEventListener("waiting",(function(){console.log("player event: waiting"),document.getElementById("log").innerHTML+="["+Date.now()+"] waiting  "})),n.addEventListener("play",(function(){console.log("player event: play"),document.getElementById("log").innerHTML+="["+Date.now()+"] play  "})),n.addEventListener("pause",(function(){console.log("player event: pause"),document.getElementById("log").innerHTML+="["+Date.now()+"] pause  "})),n.addEventListener("ended",(function(){console.log("player event: ended"),document.getElementById("log").innerHTML+="["+Date.now()+"] ended  "})),n.addEventListener("loadstart",(function(){console.log("player event: loadstart"),document.getElementById("log").innerHTML+="["+Date.now()+"] loadstart  "})),n.addEventListener("durationchange",(function(){console.log("player event: durationchange"),document.getElementById("log").innerHTML+="["+Date.now()+"] durationchange  "})),n.addEventListener("loadedmetadata",(function(){console.log("player event: loadedmetadata"),document.getElementById("log").innerHTML+="["+Date.now()+"] loadedmetadata  "})),n.addEventListener("loadeddata",(function(){console.log("player event: loadeddata"),document.getElementById("log").innerHTML+="["+Date.now()+"] loadeddata  "})),n.addEventListener("canplay",(function(){console.log("player event: canplay"),document.getElementById("log").innerHTML+="["+Date.now()+"] canplay  ",n.play()})),n.addEventListener("canplaythrough",(function(){console.log("player event: canplaythrough"),document.getElementById("log").innerHTML+="["+Date.now()+"] canplaythrough  "})),l.a.post(a).then((function(a){e.response_from_server=a.data,e.audio_src=e.audio_src_org+"/"+e.response_from_server+".wav",console.log("SID data "+e.audio_src),n.load()}))}}},i=d,c=(n("034f"),n("2877")),u=Object(c["a"])(i,r,o,!1,null,null,null),p=u.exports;l.a.defaults.headers.post["Content-Type"]="application/x-www-form-urlencoded",t["a"].config.productionTip=!1,new t["a"]({render:function(e){return e(p)},axios:l.a}).$mount("#app")},"64a9":function(e,a,n){}});
//# sourceMappingURL=app.21a472de.js.map