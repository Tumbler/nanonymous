(function(f){if(typeof exports==="object"&&typeof module!=="undefined"){module.exports=f()}else if(typeof define==="function"&&define.amd){define([],f)}else{var g;if(typeof window!=="undefined"){g=window}else if(typeof global!=="undefined"){g=global}else if(typeof self!=="undefined"){g=self}else{g=this}g.nanocurrency = f()}})(function(){var define,module,exports;return (function(){function r(e,n,t){function o(i,f){if(!n[i]){if(!e[i]){var c="function"==typeof require&&require;if(!f&&c)return c(i,!0);if(u)return u(i,!0);var a=new Error("Cannot find module '"+i+"'");throw a.code="MODULE_NOT_FOUND",a}var p=n[i]={exports:{}};e[i][0].call(p.exports,function(r){var n=e[i][1][r];return o(n||r)},p,p.exports,r,e,n,t)}return n[i].exports}for(var u="function"==typeof require&&require,i=0;i<t.length;i++)o(t[i]);return o}return r})()({1:[function(require,module,exports){
(function (process,Buffer,__dirname){(function (){
/*!
* nanocurrency-js v2.5.0: A toolkit for the Nano cryptocurrency.
* Copyright (c) 2020 Marvin ROGER <dev at marvinroger dot fr>
* Licensed under GPL-3.0 (https://git.io/vAZsK)
*/
!function(A,I){"object"==typeof exports&&"undefined"!=typeof module?I(exports,require("fs"),require("path")):"function"==typeof define&&define.amd?define(["exports","fs","path"],I):I((A=A||self).NanoCurrency={},A.fs,A.path)}(this,(function(A,I,i){"use strict";
/*! *****************************************************************************
    Copyright (c) Microsoft Corporation. All rights reserved.
    Licensed under the Apache License, Version 2.0 (the "License"); you may not use
    this file except in compliance with the License. You may obtain a copy of the
    License at http://www.apache.org/licenses/LICENSE-2.0

    THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT LIMITATION ANY IMPLIED
    WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR A PARTICULAR PURPOSE,
    MERCHANTABLITY OR NON-INFRINGEMENT.

    See the Apache Version 2.0 License for specific language governing permissions
    and limitations under the License.
    ***************************************************************************** */
function e(A,I,i,e){return new(i||(i=Promise))((function(r,n){function g(A){try{o(e.next(A))}catch(A){n(A)}}function t(A){try{o(e.throw(A))}catch(A){n(A)}}function o(A){var I;A.done?r(A.value):(I=A.value,I instanceof i?I:new i((function(A){A(I)}))).then(g,t)}o((e=e.apply(A,I||[])).next())}))}function r(A,I){var i,e,r,n,g={label:0,sent:function(){if(1&r[0])throw r[1];return r[1]},trys:[],ops:[]};return n={next:t(0),throw:t(1),return:t(2)},"function"==typeof Symbol&&(n[Symbol.iterator]=function(){return this}),n;function t(n){return function(t){return function(n){if(i)throw new TypeError("Generator is already executing.");for(;g;)try{if(i=1,e&&(r=2&n[0]?e.return:n[0]?e.throw||((r=e.return)&&r.call(e),0):e.next)&&!(r=r.call(e,n[1])).done)return r;switch(e=0,r&&(n=[2&n[0],r.value]),n[0]){case 0:case 1:r=n;break;case 4:return g.label++,{value:n[1],done:!1};case 5:g.label++,e=n[1],n=[0];continue;case 7:n=g.ops.pop(),g.trys.pop();continue;default:if(!(r=(r=g.trys).length>0&&r[r.length-1])&&(6===n[0]||2===n[0])){g=0;continue}if(3===n[0]&&(!r||n[1]>r[0]&&n[1]<r[3])){g.label=n[1];break}if(6===n[0]&&g.label<r[1]){g.label=r[1],r=n;break}if(r&&g.label<r[2]){g.label=r[2],g.ops.push(n);break}r[2]&&g.ops.pop(),g.trys.pop();continue}n=I.call(A,g)}catch(A){n=[6,A],e=0}finally{i=r=0}if(5&n[0])throw n[1];return{value:n[0]?n[1]:void 0,done:!0}}([n,t])}}}I=I&&I.hasOwnProperty("default")?I.default:I,i=i&&i.hasOwnProperty("default")?i.default:i;var n=function(A,I){return A(I={exports:{}},I.exports),I.exports}((function(A,e){var r,n=(r="undefined"!=typeof document&&document.currentScript?document.currentScript.src:void 0,function(A){var e;A=A||{},e||(e=void 0!==A?A:{});var n,g={};for(n in e)e.hasOwnProperty(n)&&(g[n]=e[n]);e.arguments=[],e.thisProgram="./this.program",e.quit=function(A,I){throw I},e.preRun=[],e.postRun=[];var t=!1,o=!1,C=!1,a=!1;t="object"==typeof window,o="function"==typeof importScripts,C="object"==typeof process&&!t&&!o,a=!t&&!C&&!o;var h,f,u="";C?(u=__dirname+"/",e.read=function(A,e){var r=T(A);return r||(h||(h=I),f||(f=i),A=f.normalize(A),r=h.readFileSync(A)),e?r:r.toString()},e.readBinary=function(A){return(A=e.read(A,!0)).buffer||(A=new Uint8Array(A)),l(A.buffer),A},1<process.argv.length&&(e.thisProgram=process.argv[1].replace(/\\/g,"/")),e.arguments=process.argv.slice(2),process.on("uncaughtException",(function(A){if(!(A instanceof _))throw A})),process.on("unhandledRejection",$),e.quit=function(A){process.exit(A)},e.inspect=function(){return"[Emscripten Module object]"}):a?("undefined"!=typeof read&&(e.read=function(A){var I=T(A);return I?J(I):read(A)}),e.readBinary=function(A){var I;return(I=T(A))?I:"function"==typeof readbuffer?new Uint8Array(readbuffer(A)):(l("object"==typeof(I=read(A,"binary"))),I)},"undefined"!=typeof scriptArgs?e.arguments=scriptArgs:void 0!==arguments&&(e.arguments=arguments),"function"==typeof quit&&(e.quit=function(A){quit(A)})):(t||o)&&(o?u=self.location.href:document.currentScript&&(u=document.currentScript.src),r&&(u=r),u=0!==u.indexOf("blob:")?u.substr(0,u.lastIndexOf("/")+1):"",e.read=function(A){try{var I=new XMLHttpRequest;return I.open("GET",A,!1),I.send(null),I.responseText}catch(I){if(A=T(A))return J(A);throw I}},o&&(e.readBinary=function(A){try{var I=new XMLHttpRequest;return I.open("GET",A,!1),I.responseType="arraybuffer",I.send(null),new Uint8Array(I.response)}catch(I){if(A=T(A))return A;throw I}}),e.readAsync=function(A,I,i){var e=new XMLHttpRequest;e.open("GET",A,!0),e.responseType="arraybuffer",e.onload=function(){if(200==e.status||0==e.status&&e.response)I(e.response);else{var r=T(A);r?I(r.buffer):i()}},e.onerror=i,e.send(null)},e.setWindowTitle=function(A){document.title=A});var s=e.print||("undefined"!=typeof console?console.log.bind(console):"undefined"!=typeof print?print:null),Q=e.printErr||("undefined"!=typeof printErr?printErr:"undefined"!=typeof console&&console.warn.bind(console)||s);for(n in g)g.hasOwnProperty(n)&&(e[n]=g[n]);g=void 0;var B={"f64-rem":function(A,I){return A%I},debugger:function(){}};"object"!=typeof WebAssembly&&Q("no native wasm support detected");var E,c=!1;function l(A,I){A||$("Assertion failed: "+I)}function w(A){var I=e["_"+A];return l(I,"Cannot call unknown function "+A+", make sure it is exported"),I}function U(A,I,i,e){var r={string:function(A){var I=0;if(null!=A&&0!==A){var i=1+(A.length<<2),e=I=q(i),r=d;if(0<i){i=e+i-1;for(var n=0;n<A.length;++n){var g=A.charCodeAt(n);if(55296<=g&&57343>=g&&(g=65536+((1023&g)<<10)|1023&A.charCodeAt(++n)),127>=g){if(e>=i)break;r[e++]=g}else{if(2047>=g){if(e+1>=i)break;r[e++]=192|g>>6}else{if(65535>=g){if(e+2>=i)break;r[e++]=224|g>>12}else{if(e+3>=i)break;r[e++]=240|g>>18,r[e++]=128|g>>12&63}r[e++]=128|g>>6&63}r[e++]=128|63&g}}r[e]=0}}return I},array:function(A){var I=q(A.length);return y.set(A,I),I}},n=w(A),g=[];if(A=0,e)for(var t=0;t<e.length;t++){var o=r[i[t]];o?(0===A&&(A=W()),g[t]=o(e[t])):g[t]=e[t]}return i=function(A){if("string"===I)if(A){for(var i=d,e=A+void 0,r=A;i[r]&&!(r>=e);)++r;if(16<r-A&&i.subarray&&S)A=S.decode(i.subarray(A,r));else{for(e="";A<r;){var n=i[A++];if(128&n){var g=63&i[A++];if(192==(224&n))e+=String.fromCharCode((31&n)<<6|g);else{var t=63&i[A++];65536>(n=224==(240&n)?(15&n)<<12|g<<6|t:(7&n)<<18|g<<12|t<<6|63&i[A++])?e+=String.fromCharCode(n):(n-=65536,e+=String.fromCharCode(55296|n>>10,56320|1023&n))}}else e+=String.fromCharCode(n)}A=e}}else A="";else A="boolean"===I?!!A:A;return A}(i=n.apply(null,g)),0!==A&&Z(A),i}var S="undefined"!=typeof TextDecoder?new TextDecoder("utf8"):void 0;"undefined"!=typeof TextDecoder&&new TextDecoder("utf-16le");var F,y,d,p,G=e.TOTAL_MEMORY||16777216;function D(A){for(;0<A.length;){var I=A.shift();if("function"==typeof I)I();else{var i=I.h;"number"==typeof i?void 0===I.g?e.dynCall_v(i):e.dynCall_vi(i,I.g):i(void 0===I.g?null:I.g)}}}5242880>G&&Q("TOTAL_MEMORY should be larger than TOTAL_STACK, was "+G+"! (TOTAL_STACK=5242880)"),e.buffer?F=e.buffer:(F="object"==typeof WebAssembly&&"function"==typeof WebAssembly.Memory?(E=new WebAssembly.Memory({initial:G/65536,maximum:G/65536})).buffer:new ArrayBuffer(G),e.buffer=F),e.HEAP8=y=new Int8Array(F),e.HEAP16=new Int16Array(F),e.HEAP32=p=new Int32Array(F),e.HEAPU8=d=new Uint8Array(F),e.HEAPU16=new Uint16Array(F),e.HEAPU32=new Uint32Array(F),e.HEAPF32=new Float32Array(F),e.HEAPF64=new Float64Array(F),p[724]=5246032;var v=[],k=[],b=[],H=[],m=!1;function M(){var A=e.preRun.shift();v.unshift(A)}var Y=0,N=null;e.preloadedImages={},e.preloadedAudios={};var x="data:application/octet-stream;base64,";function K(A){return String.prototype.startsWith?A.startsWith(x):0===A.indexOf(x)}var R="data:application/octet-stream;base64,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NCAAIAM2AjggACADNgI8IABBQGshAAwBCwsDQCAAIAVIBEAgACADNgIAIABBBGohAAwBCwsLA0AgACAESARAIAAgAToAACAAQQFqIQAMAQsLIAQgAmsLxgMBA38gAkGAwABOBEAgACABIAIQABogAA8LIAAhBCAAIAJqIQMgAEEDcSABQQNxRgRAA0AgAEEDcQRAIAJFBEAgBA8LIAAgASwAADoAACAAQQFqIQAgAUEBaiEBIAJBAWshAgwBCwsgA0F8cSICQUBqIQUDQCAAIAVMBEAgACABKAIANgIAIAAgASgCBDYCBCAAIAEoAgg2AgggACABKAIMNgIMIAAgASgCEDYCECAAIAEoAhQ2AhQgACABKAIYNgIYIAAgASgCHDYCHCAAIAEoAiA2AiAgACABKAIkNgIkIAAgASgCKDYCKCAAIAEoAiw2AiwgACABKAIwNgIwIAAgASgCNDYCNCAAIAEoAjg2AjggACABKAI8NgI8IABBQGshACABQUBrIQEMAQsLA0AgACACSARAIAAgASgCADYCACAAQQRqIQAgAUEEaiEBDAELCwUgA0EEayECA0AgACACSARAIAAgASwAADoAACAAIAEsAAE6AAEgACABLAACOgACIAAgASwAAzoAAyAAQQRqIQAgAUEEaiEBDAELCwsDQCAAIANIBEAgACABLAAAOgAAIABBAWohACABQQFqIQEMAQsLIAQLBwAgABAMpwuNAQEDfwJAAkAgACICQQNxRQ0AIAIiASEAAkADQCABLAAARQ0BIAFBAWoiASIAQQNxDQALIAEhAAwBCwwBCwNAIABBBGohASAAKAIAIgNB//37d2ogA0GAgYKEeHFBgIGChHhzcUUEQCABIQAMAQsLIANB/wFxBEADQCAAQQFqIgAsAAANAAsLCyAAIAJrC+wFAgR/AX4DQCAAKAIEIgEgACgCZEkEfyAAIAFBAWo2AgQgAS0AAAUgABACCyIBIgNBIEYgA0F3akEFSXINAAsCQAJAIAFBK2sOAwABAAELIAFBLUZBH3RBH3UhBCAAKAIEIgEgACgCZEkEfyAAIAFBAWo2AgQgAS0AAAUgABACCyEBCwJ+An8CQAJAIAFBMEYEfiAAKAIEIgEgACgCZEkEfyAAIAFBAWo2AgQgAS0AAAUgABACCyIBQSByQfgARwRAIAFBkQhqLAAAIgNB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aiwAADoAAEG2CiABQf8BcUEEdkGACGosAAA6AABBtwogAUEPcUGACGosAAA6AABBuAogAkH/AXFBBHZBgAhqLAAAOgAAQbkKIAJBD3FBgAhqLAAAOgAAQboKIANB/wFxQQR2QYAIaiwAADoAAEG7CiADQQ9xQYAIaiwAADoAAEG8CiAJQf8BcUEEdkGACGosAAA6AABBvQogCUEPcUGACGosAAA6AABBvgogBUH/AXFBBHZBgAhqLAAAOgAAQb8KIAVBD3FBgAhqLAAAOgAAQcAKIAhB/wFxQQR2QYAIaiwAADoAAEHBCiAIQQ9xQYAIaiwAADoAAEHCCkEAOgAAIAckAUGwCgsGACAAJAELCABBABABQQALcAIBfwJ+IwEhASMBQYABaiQBIAFBADYCACABIAA2AgQgASAANgIsIAFBfyAAQf////8HaiAAQQBIGzYCCCABQX82AkwgAUEANgJoIAEgASgCCCIAIAEoAgRrNgJsIAEgADYCZCABEAghAyABJAEgAwuLAQECfyAAIAAsAEoiASABQf8BanI6AEogACgCFCAAKAIcSwRAIAAoAiQhASAAQQBBACABQQFxEQAAGgsgAEEANgIQIABBADYCHCAAQQA2AhQgACgCACIBQQRxBH8gACABQSByNgIAQX8FIAAgACgCLCAAKAIwaiICNgIIIAAgAjYCBCABQRt0QR91CwtEAQN/IwEhASMBQRBqJAEgABANBH9BfwUgACgCICECIAAgAUEBIAJBAXERAABBAUYEfyABLQAABUF/CwshAyABJAEgAwsEACMBCxsBAn8jASECIAAjAWokASMBQQ9qQXBxJAEgAgsLoAICAEGACAuRAjAxMjM0NTY3ODlhYmNkZWb/////////////////////////////////////////////////////////////////AAECAwQFBgcICf////////8KCwwNDg8QERITFBUWFxgZGhscHR4fICEiI////////woLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIj/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////wBBoAoLAQE=";if(!K(R)){var L=R;R=e.locateFile?e.locateFile(L,u):u+L}function O(){try{if(e.wasmBinary)return new Uint8Array(e.wasmBinary);var A=T(R);if(A)return A;if(e.readBinary)return e.readBinary(R);throw"both async and sync fetching of the wasm failed"}catch(A){$(A)}}function P(){return e.wasmBinary||!t&&!o||"function"!=typeof fetch?new Promise((function(A){A(O())})):fetch(R,{credentials:"same-origin"}).then((function(A){if(!A.ok)throw"failed to load wasm binary file at '"+R+"'";return A.arrayBuffer()})).catch((function(){return O()}))}function j(A){function I(A){e.asm=A.exports,Y--,e.monitorRunDependencies&&e.monitorRunDependencies(Y),0==Y&&N&&(A=N,N=null,A())}function i(A){I(A.instance)}function r(A){P().then((function(A){return WebAssembly.instantiate(A,n)})).then(A,(function(A){Q("failed to asynchronously prepare wasm: "+A),$(A)}))}var n={env:A,global:{NaN:NaN,Infinity:1/0},"global.Math":Math,asm2wasm:B};if(Y++,e.monitorRunDependencies&&e.monitorRunDependencies(Y),e.instantiateWasm)try{return e.instantiateWasm(n,I)}catch(A){return Q("Module.instantiateWasm callback failed with error: "+A),!1}return e.wasmBinary||"function"!=typeof WebAssembly.instantiateStreaming||K(R)||"function"!=typeof fetch?r(i):WebAssembly.instantiateStreaming(fetch(R,{credentials:"same-origin"}),n).then(i,(function(A){Q("wasm streaming compile failed: "+A),Q("falling back to ArrayBuffer instantiation"),r(i)})),{}}function J(A){for(var I=[],i=0;i<A.length;i++){var e=A[i];255<e&&(e&=255),I.push(String.fromCharCode(e))}return I.join("")}e.asm=function(A,I){return I.memory=E,I.table=new WebAssembly.Table({initial:2,maximum:2,element:"anyfunc"}),I.__memory_base=1024,I.__table_base=0,j(I)};var X="function"==typeof atob?atob:function(A){var I="",i=0;A=A.replace(/[^A-Za-z0-9\+\/=]/g,"");do{var e="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=".indexOf(A.charAt(i++)),r="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=".indexOf(A.charAt(i++)),n="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=".indexOf(A.charAt(i++)),g="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=".indexOf(A.charAt(i++));e=e<<2|r>>4,r=(15&r)<<4|n>>2;var t=(3&n)<<6|g;I+=String.fromCharCode(e),64!==n&&(I+=String.fromCharCode(r)),64!==g&&(I+=String.fromCharCode(t))}while(i<A.length);return I};function T(A){if(K(A)){if(A=A.slice(x.length),"boolean"==typeof C&&C){try{var I=Buffer.from(A,"base64")}catch(i){I=new Buffer(A,"base64")}var i=new Uint8Array(I.buffer,I.byteOffset,I.byteLength)}else try{var e=X(A),r=new Uint8Array(e.length);for(I=0;I<e.length;++I)r[I]=e.charCodeAt(I);i=r}catch(A){throw Error("Converting base64 string to bytes failed.")}return i}}var V=e.asm({},{b:$,a:function(A,I,i){d.set(d.subarray(I,I+i),A)}},F);e.asm=V,e._emscripten_work=function(){return e.asm.c.apply(null,arguments)};var q=e.stackAlloc=function(){return e.asm.d.apply(null,arguments)},Z=e.stackRestore=function(){return e.asm.e.apply(null,arguments)},W=e.stackSave=function(){return e.asm.f.apply(null,arguments)};function _(A){this.name="ExitStatus",this.message="Program terminated with exit("+A+")",this.status=A}function z(){function A(){if(!e.calledRun&&(e.calledRun=!0,!c)){if(m||(m=!0,D(k)),D(b),e.onRuntimeInitialized&&e.onRuntimeInitialized(),e.postRun)for("function"==typeof e.postRun&&(e.postRun=[e.postRun]);e.postRun.length;){var A=e.postRun.shift();H.unshift(A)}D(H)}}if(!(0<Y)){if(e.preRun)for("function"==typeof e.preRun&&(e.preRun=[e.preRun]);e.preRun.length;)M();D(v),0<Y||e.calledRun||(e.setStatus?(e.setStatus("Running..."),setTimeout((function(){setTimeout((function(){e.setStatus("")}),1),A()}),1)):A())}}function $(A){throw e.onAbort&&e.onAbort(A),void 0!==A?(s(A),Q(A),A=JSON.stringify(A)):A="",c=!0,"abort("+A+"). Build with -s ASSERTIONS=1 for more info."}if(e.asm=V,e.cwrap=function(A,I,i,e){var r=(i=i||[]).every((function(A){return"number"===A}));return"string"!==I&&r&&!e?w(A):function(){return U(A,I,i,arguments)}},e.then=function(A){if(e.calledRun)A(e);else{var I=e.onRuntimeInitialized;e.onRuntimeInitialized=function(){I&&I(),A(e)}}return e},_.prototype=Error(),_.prototype.constructor=_,N=function A(){e.calledRun||z(),e.calledRun||(N=A)},e.run=z,e.abort=$,e.preInit)for("function"==typeof e.preInit&&(e.preInit=[e.preInit]);0<e.preInit.length;)e.preInit.pop()();return e.noExitRuntime=!0,z(),A});A.exports=n})),g=/^-?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?$/i,t=Math.ceil,o=Math.floor,C="[BigNumber Error] ",a=C+"Number primitive has more than 15 significant digits: ",h=[1,10,100,1e3,1e4,1e5,1e6,1e7,1e8,1e9,1e10,1e11,1e12,1e13];function f(A){var I=0|A;return A>0||A===I?I:I-1}function u(A){for(var I,i,e=1,r=A.length,n=A[0]+"";e<r;){for(i=14-(I=A[e++]+"").length;i--;I="0"+I);n+=I}for(r=n.length;48===n.charCodeAt(--r););return n.slice(0,r+1||1)}function s(A,I){var i,e,r=A.c,n=I.c,g=A.s,t=I.s,o=A.e,C=I.e;if(!g||!t)return null;if(i=r&&!r[0],e=n&&!n[0],i||e)return i?e?0:-t:g;if(g!=t)return g;if(i=g<0,e=o==C,!r||!n)return e?0:!r^i?1:-1;if(!e)return o>C^i?1:-1;for(t=(o=r.length)<(C=n.length)?o:C,g=0;g<t;g++)if(r[g]!=n[g])return r[g]>n[g]^i?1:-1;return o==C?0:o>C^i?1:-1}function Q(A,I,i,e){if(A<I||A>i||A!==o(A))throw Error(C+(e||"Argument")+("number"==typeof A?A<I||A>i?" out of range: ":" not an integer: ":" not a primitive number: ")+String(A))}function B(A){var I=A.c.length-1;return f(A.e/14)==I&&A.c[I]%2!=0}function E(A,I){return(A.length>1?A.charAt(0)+"."+A.slice(1):A)+(I<0?"e":"e+")+I}function c(A,I,i){var e,r;if(I<0){for(r=i+".";++I;r+=i);A=r+A}else if(++I>(e=A.length)){for(r=i,I-=e;--I;r+=i);A+=r}else I<e&&(A=A.slice(0,I)+"."+A.slice(I));return A}var l=function A(I){var i,e,r,n,l,w,U,S,F,y=x.prototype={constructor:x,toString:null,valueOf:null},d=new x(1),p=20,G=4,D=-7,v=21,k=-1e7,b=1e7,H=!1,m=1,M=0,Y={prefix:"",groupSize:3,secondaryGroupSize:0,groupSeparator:",",decimalSeparator:".",fractionGroupSize:0,fractionGroupSeparator:" ",suffix:""},N="0123456789abcdefghijklmnopqrstuvwxyz";function x(A,I){var i,n,t,C,h,f,u,s,B=this;if(!(B instanceof x))return new x(A,I);if(null==I){if(A&&!0===A._isBigNumber)return B.s=A.s,void(!A.c||A.e>b?B.c=B.e=null:A.e<k?B.c=[B.e=0]:(B.e=A.e,B.c=A.c.slice()));if((f="number"==typeof A)&&0*A==0){if(B.s=1/A<0?(A=-A,-1):1,A===~~A){for(C=0,h=A;h>=10;h/=10,C++);return void(C>b?B.c=B.e=null:(B.e=C,B.c=[A]))}s=String(A)}else{if(!g.test(s=String(A)))return r(B,s,f);B.s=45==s.charCodeAt(0)?(s=s.slice(1),-1):1}(C=s.indexOf("."))>-1&&(s=s.replace(".","")),(h=s.search(/e/i))>0?(C<0&&(C=h),C+=+s.slice(h+1),s=s.substring(0,h)):C<0&&(C=s.length)}else{if(Q(I,2,N.length,"Base"),10==I)return O(B=new x(A),p+B.e+1,G);if(s=String(A),f="number"==typeof A){if(0*A!=0)return r(B,s,f,I);if(B.s=1/A<0?(s=s.slice(1),-1):1,x.DEBUG&&s.replace(/^0\.0*|\./,"").length>15)throw Error(a+A)}else B.s=45===s.charCodeAt(0)?(s=s.slice(1),-1):1;for(i=N.slice(0,I),C=h=0,u=s.length;h<u;h++)if(i.indexOf(n=s.charAt(h))<0){if("."==n){if(h>C){C=u;continue}}else if(!t&&(s==s.toUpperCase()&&(s=s.toLowerCase())||s==s.toLowerCase()&&(s=s.toUpperCase()))){t=!0,h=-1,C=0;continue}return r(B,String(A),f,I)}f=!1,(C=(s=e(s,I,10,B.s)).indexOf("."))>-1?s=s.replace(".",""):C=s.length}for(h=0;48===s.charCodeAt(h);h++);for(u=s.length;48===s.charCodeAt(--u););if(s=s.slice(h,++u)){if(u-=h,f&&x.DEBUG&&u>15&&(A>9007199254740991||A!==o(A)))throw Error(a+B.s*A);if((C=C-h-1)>b)B.c=B.e=null;else if(C<k)B.c=[B.e=0];else{if(B.e=C,B.c=[],h=(C+1)%14,C<0&&(h+=14),h<u){for(h&&B.c.push(+s.slice(0,h)),u-=14;h<u;)B.c.push(+s.slice(h,h+=14));h=14-(s=s.slice(h)).length}else h-=u;for(;h--;s+="0");B.c.push(+s)}}else B.c=[B.e=0]}function K(A,I,i,e){var r,n,g,t,o;if(null==i?i=G:Q(i,0,8),!A.c)return A.toString();if(r=A.c[0],g=A.e,null==I)o=u(A.c),o=1==e||2==e&&(g<=D||g>=v)?E(o,g):c(o,g,"0");else if(n=(A=O(new x(A),I,i)).e,t=(o=u(A.c)).length,1==e||2==e&&(I<=n||n<=D)){for(;t<I;o+="0",t++);o=E(o,n)}else if(I-=g,o=c(o,n,"0"),n+1>t){if(--I>0)for(o+=".";I--;o+="0");}else if((I+=n-t)>0)for(n+1==t&&(o+=".");I--;o+="0");return A.s<0&&r?"-"+o:o}function R(A,I){for(var i,e=1,r=new x(A[0]);e<A.length;e++){if(!(i=new x(A[e])).s){r=i;break}I.call(r,i)&&(r=i)}return r}function L(A,I,i){for(var e=1,r=I.length;!I[--r];I.pop());for(r=I[0];r>=10;r/=10,e++);return(i=e+14*i-1)>b?A.c=A.e=null:i<k?A.c=[A.e=0]:(A.e=i,A.c=I),A}function O(A,I,i,e){var r,n,g,C,a,f,u,s=A.c,Q=h;if(s){A:{for(r=1,C=s[0];C>=10;C/=10,r++);if((n=I-r)<0)n+=14,g=I,u=(a=s[f=0])/Q[r-g-1]%10|0;else if((f=t((n+1)/14))>=s.length){if(!e)break A;for(;s.length<=f;s.push(0));a=u=0,r=1,g=(n%=14)-14+1}else{for(a=C=s[f],r=1;C>=10;C/=10,r++);u=(g=(n%=14)-14+r)<0?0:a/Q[r-g-1]%10|0}if(e=e||I<0||null!=s[f+1]||(g<0?a:a%Q[r-g-1]),e=i<4?(u||e)&&(0==i||i==(A.s<0?3:2)):u>5||5==u&&(4==i||e||6==i&&(n>0?g>0?a/Q[r-g]:0:s[f-1])%10&1||i==(A.s<0?8:7)),I<1||!s[0])return s.length=0,e?(I-=A.e+1,s[0]=Q[(14-I%14)%14],A.e=-I||0):s[0]=A.e=0,A;if(0==n?(s.length=f,C=1,f--):(s.length=f+1,C=Q[14-n],s[f]=g>0?o(a/Q[r-g]%Q[g])*C:0),e)for(;;){if(0==f){for(n=1,g=s[0];g>=10;g/=10,n++);for(g=s[0]+=C,C=1;g>=10;g/=10,C++);n!=C&&(A.e++,1e14==s[0]&&(s[0]=1));break}if(s[f]+=C,1e14!=s[f])break;s[f--]=0,C=1}for(n=s.length;0===s[--n];s.pop());}A.e>b?A.c=A.e=null:A.e<k&&(A.c=[A.e=0])}return A}function P(A){var I,i=A.e;return null===i?A.toString():(I=u(A.c),I=i<=D||i>=v?E(I,i):c(I,i,"0"),A.s<0?"-"+I:I)}return x.clone=A,x.ROUND_UP=0,x.ROUND_DOWN=1,x.ROUND_CEIL=2,x.ROUND_FLOOR=3,x.ROUND_HALF_UP=4,x.ROUND_HALF_DOWN=5,x.ROUND_HALF_EVEN=6,x.ROUND_HALF_CEIL=7,x.ROUND_HALF_FLOOR=8,x.EUCLID=9,x.config=x.set=function(A){var I,i;if(null!=A){if("object"!=typeof A)throw Error(C+"Object expected: "+A);if(A.hasOwnProperty(I="DECIMAL_PLACES")&&(Q(i=A[I],0,1e9,I),p=i),A.hasOwnProperty(I="ROUNDING_MODE")&&(Q(i=A[I],0,8,I),G=i),A.hasOwnProperty(I="EXPONENTIAL_AT")&&((i=A[I])&&i.pop?(Q(i[0],-1e9,0,I),Q(i[1],0,1e9,I),D=i[0],v=i[1]):(Q(i,-1e9,1e9,I),D=-(v=i<0?-i:i))),A.hasOwnProperty(I="RANGE"))if((i=A[I])&&i.pop)Q(i[0],-1e9,-1,I),Q(i[1],1,1e9,I),k=i[0],b=i[1];else{if(Q(i,-1e9,1e9,I),!i)throw Error(C+I+" cannot be zero: "+i);k=-(b=i<0?-i:i)}if(A.hasOwnProperty(I="CRYPTO")){if((i=A[I])!==!!i)throw Error(C+I+" not true or false: "+i);if(i){if("undefined"==typeof crypto||!crypto||!crypto.getRandomValues&&!crypto.randomBytes)throw H=!i,Error(C+"crypto unavailable");H=i}else H=i}if(A.hasOwnProperty(I="MODULO_MODE")&&(Q(i=A[I],0,9,I),m=i),A.hasOwnProperty(I="POW_PRECISION")&&(Q(i=A[I],0,1e9,I),M=i),A.hasOwnProperty(I="FORMAT")){if("object"!=typeof(i=A[I]))throw Error(C+I+" not an object: "+i);Y=i}if(A.hasOwnProperty(I="ALPHABET")){if("string"!=typeof(i=A[I])||/^.$|[+-.\s]|(.).*\1/.test(i))throw Error(C+I+" invalid: "+i);N=i}}return{DECIMAL_PLACES:p,ROUNDING_MODE:G,EXPONENTIAL_AT:[D,v],RANGE:[k,b],CRYPTO:H,MODULO_MODE:m,POW_PRECISION:M,FORMAT:Y,ALPHABET:N}},x.isBigNumber=function(A){if(!A||!0!==A._isBigNumber)return!1;if(!x.DEBUG)return!0;var I,i,e=A.c,r=A.e,n=A.s;A:if("[object Array]"=={}.toString.call(e)){if((1===n||-1===n)&&r>=-1e9&&r<=1e9&&r===o(r)){if(0===e[0]){if(0===r&&1===e.length)return!0;break A}if((I=(r+1)%14)<1&&(I+=14),String(e[0]).length==I){for(I=0;I<e.length;I++)if((i=e[I])<0||i>=1e14||i!==o(i))break A;if(0!==i)return!0}}}else if(null===e&&null===r&&(null===n||1===n||-1===n))return!0;throw Error(C+"Invalid BigNumber: "+A)},x.maximum=x.max=function(){return R(arguments,y.lt)},x.minimum=x.min=function(){return R(arguments,y.gt)},x.random=(n=9007199254740992*Math.random()&2097151?function(){return o(9007199254740992*Math.random())}:function(){return 8388608*(1073741824*Math.random()|0)+(8388608*Math.random()|0)},function(A){var I,i,e,r,g,a=0,f=[],u=new x(d);if(null==A?A=p:Q(A,0,1e9),r=t(A/14),H)if(crypto.getRandomValues){for(I=crypto.getRandomValues(new Uint32Array(r*=2));a<r;)(g=131072*I[a]+(I[a+1]>>>11))>=9e15?(i=crypto.getRandomValues(new Uint32Array(2)),I[a]=i[0],I[a+1]=i[1]):(f.push(g%1e14),a+=2);a=r/2}else{if(!crypto.randomBytes)throw H=!1,Error(C+"crypto unavailable");for(I=crypto.randomBytes(r*=7);a<r;)(g=281474976710656*(31&I[a])+1099511627776*I[a+1]+4294967296*I[a+2]+16777216*I[a+3]+(I[a+4]<<16)+(I[a+5]<<8)+I[a+6])>=9e15?crypto.randomBytes(7).copy(I,a):(f.push(g%1e14),a+=7);a=r/7}if(!H)for(;a<r;)(g=n())<9e15&&(f[a++]=g%1e14);for(A%=14,(r=f[--a])&&A&&(g=h[14-A],f[a]=o(r/g)*g);0===f[a];f.pop(),a--);if(a<0)f=[e=0];else{for(e=-1;0===f[0];f.splice(0,1),e-=14);for(a=1,g=f[0];g>=10;g/=10,a++);a<14&&(e-=14-a)}return u.e=e,u.c=f,u}),x.sum=function(){for(var A=1,I=arguments,i=new x(I[0]);A<I.length;)i=i.plus(I[A++]);return i},e=function(){function A(A,I,i,e){for(var r,n,g=[0],t=0,o=A.length;t<o;){for(n=g.length;n--;g[n]*=I);for(g[0]+=e.indexOf(A.charAt(t++)),r=0;r<g.length;r++)g[r]>i-1&&(null==g[r+1]&&(g[r+1]=0),g[r+1]+=g[r]/i|0,g[r]%=i)}return g.reverse()}return function(I,e,r,n,g){var t,o,C,a,h,f,s,Q,B=I.indexOf("."),E=p,l=G;for(B>=0&&(a=M,M=0,I=I.replace(".",""),f=(Q=new x(e)).pow(I.length-B),M=a,Q.c=A(c(u(f.c),f.e,"0"),10,r,"0123456789"),Q.e=Q.c.length),C=a=(s=A(I,e,r,g?(t=N,"0123456789"):(t="0123456789",N))).length;0==s[--a];s.pop());if(!s[0])return t.charAt(0);if(B<0?--C:(f.c=s,f.e=C,f.s=n,s=(f=i(f,Q,E,l,r)).c,h=f.r,C=f.e),B=s[o=C+E+1],a=r/2,h=h||o<0||null!=s[o+1],h=l<4?(null!=B||h)&&(0==l||l==(f.s<0?3:2)):B>a||B==a&&(4==l||h||6==l&&1&s[o-1]||l==(f.s<0?8:7)),o<1||!s[0])I=h?c(t.charAt(1),-E,t.charAt(0)):t.charAt(0);else{if(s.length=o,h)for(--r;++s[--o]>r;)s[o]=0,o||(++C,s=[1].concat(s));for(a=s.length;!s[--a];);for(B=0,I="";B<=a;I+=t.charAt(s[B++]));I=c(I,C,t.charAt(0))}return I}}(),i=function(){function A(A,I,i){var e,r,n,g,t=0,o=A.length,C=I%1e7,a=I/1e7|0;for(A=A.slice();o--;)t=((r=C*(n=A[o]%1e7)+(e=a*n+(g=A[o]/1e7|0)*C)%1e7*1e7+t)/i|0)+(e/1e7|0)+a*g,A[o]=r%i;return t&&(A=[t].concat(A)),A}function I(A,I,i,e){var r,n;if(i!=e)n=i>e?1:-1;else for(r=n=0;r<i;r++)if(A[r]!=I[r]){n=A[r]>I[r]?1:-1;break}return n}function i(A,I,i,e){for(var r=0;i--;)A[i]-=r,r=A[i]<I[i]?1:0,A[i]=r*e+A[i]-I[i];for(;!A[0]&&A.length>1;A.splice(0,1));}return function(e,r,n,g,t){var C,a,h,u,s,Q,B,E,c,l,w,U,S,F,y,d,p,G=e.s==r.s?1:-1,D=e.c,v=r.c;if(!(D&&D[0]&&v&&v[0]))return new x(e.s&&r.s&&(D?!v||D[0]!=v[0]:v)?D&&0==D[0]||!v?0*G:G/0:NaN);for(c=(E=new x(G)).c=[],G=n+(a=e.e-r.e)+1,t||(t=1e14,a=f(e.e/14)-f(r.e/14),G=G/14|0),h=0;v[h]==(D[h]||0);h++);if(v[h]>(D[h]||0)&&a--,G<0)c.push(1),u=!0;else{for(F=D.length,d=v.length,h=0,G+=2,(s=o(t/(v[0]+1)))>1&&(v=A(v,s,t),D=A(D,s,t),d=v.length,F=D.length),S=d,w=(l=D.slice(0,d)).length;w<d;l[w++]=0);p=v.slice(),p=[0].concat(p),y=v[0],v[1]>=t/2&&y++;do{if(s=0,(C=I(v,l,d,w))<0){if(U=l[0],d!=w&&(U=U*t+(l[1]||0)),(s=o(U/y))>1)for(s>=t&&(s=t-1),B=(Q=A(v,s,t)).length,w=l.length;1==I(Q,l,B,w);)s--,i(Q,d<B?p:v,B,t),B=Q.length,C=1;else 0==s&&(C=s=1),B=(Q=v.slice()).length;if(B<w&&(Q=[0].concat(Q)),i(l,Q,w,t),w=l.length,-1==C)for(;I(v,l,d,w)<1;)s++,i(l,d<w?p:v,w,t),w=l.length}else 0===C&&(s++,l=[0]);c[h++]=s,l[0]?l[w++]=D[S]||0:(l=[D[S]],w=1)}while((S++<F||null!=l[0])&&G--);u=null!=l[0],c[0]||c.splice(0,1)}if(1e14==t){for(h=1,G=c[0];G>=10;G/=10,h++);O(E,n+(E.e=h+14*a-1)+1,g,u)}else E.e=a,E.r=+u;return E}}(),l=/^(-?)0([xbo])(?=\w[\w.]*$)/i,w=/^([^.]+)\.$/,U=/^\.([^.]+)$/,S=/^-?(Infinity|NaN)$/,F=/^\s*\+(?=[\w.])|^\s+|\s+$/g,r=function(A,I,i,e){var r,n=i?I:I.replace(F,"");if(S.test(n))A.s=isNaN(n)?null:n<0?-1:1;else{if(!i&&(n=n.replace(l,(function(A,I,i){return r="x"==(i=i.toLowerCase())?16:"b"==i?2:8,e&&e!=r?A:I})),e&&(r=e,n=n.replace(w,"$1").replace(U,"0.$1")),I!=n))return new x(n,r);if(x.DEBUG)throw Error(C+"Not a"+(e?" base "+e:"")+" number: "+I);A.s=null}A.c=A.e=null},y.absoluteValue=y.abs=function(){var A=new x(this);return A.s<0&&(A.s=1),A},y.comparedTo=function(A,I){return s(this,new x(A,I))},y.decimalPlaces=y.dp=function(A,I){var i,e,r,n=this;if(null!=A)return Q(A,0,1e9),null==I?I=G:Q(I,0,8),O(new x(n),A+n.e+1,I);if(!(i=n.c))return null;if(e=14*((r=i.length-1)-f(this.e/14)),r=i[r])for(;r%10==0;r/=10,e--);return e<0&&(e=0),e},y.dividedBy=y.div=function(A,I){return i(this,new x(A,I),p,G)},y.dividedToIntegerBy=y.idiv=function(A,I){return i(this,new x(A,I),0,1)},y.exponentiatedBy=y.pow=function(A,I){var i,e,r,n,g,a,h,f,u=this;if((A=new x(A)).c&&!A.isInteger())throw Error(C+"Exponent not an integer: "+P(A));if(null!=I&&(I=new x(I)),g=A.e>14,!u.c||!u.c[0]||1==u.c[0]&&!u.e&&1==u.c.length||!A.c||!A.c[0])return f=new x(Math.pow(+P(u),g?2-B(A):+P(A))),I?f.mod(I):f;if(a=A.s<0,I){if(I.c?!I.c[0]:!I.s)return new x(NaN);(e=!a&&u.isInteger()&&I.isInteger())&&(u=u.mod(I))}else{if(A.e>9&&(u.e>0||u.e<-1||(0==u.e?u.c[0]>1||g&&u.c[1]>=24e7:u.c[0]<8e13||g&&u.c[0]<=9999975e7)))return n=u.s<0&&B(A)?-0:0,u.e>-1&&(n=1/n),new x(a?1/n:n);M&&(n=t(M/14+2))}for(g?(i=new x(.5),a&&(A.s=1),h=B(A)):h=(r=Math.abs(+P(A)))%2,f=new x(d);;){if(h){if(!(f=f.times(u)).c)break;n?f.c.length>n&&(f.c.length=n):e&&(f=f.mod(I))}if(r){if(0===(r=o(r/2)))break;h=r%2}else if(O(A=A.times(i),A.e+1,1),A.e>14)h=B(A);else{if(0===(r=+P(A)))break;h=r%2}u=u.times(u),n?u.c&&u.c.length>n&&(u.c.length=n):e&&(u=u.mod(I))}return e?f:(a&&(f=d.div(f)),I?f.mod(I):n?O(f,M,G,void 0):f)},y.integerValue=function(A){var I=new x(this);return null==A?A=G:Q(A,0,8),O(I,I.e+1,A)},y.isEqualTo=y.eq=function(A,I){return 0===s(this,new x(A,I))},y.isFinite=function(){return!!this.c},y.isGreaterThan=y.gt=function(A,I){return s(this,new x(A,I))>0},y.isGreaterThanOrEqualTo=y.gte=function(A,I){return 1===(I=s(this,new x(A,I)))||0===I},y.isInteger=function(){return!!this.c&&f(this.e/14)>this.c.length-2},y.isLessThan=y.lt=function(A,I){return s(this,new x(A,I))<0},y.isLessThanOrEqualTo=y.lte=function(A,I){return-1===(I=s(this,new x(A,I)))||0===I},y.isNaN=function(){return!this.s},y.isNegative=function(){return this.s<0},y.isPositive=function(){return this.s>0},y.isZero=function(){return!!this.c&&0==this.c[0]},y.minus=function(A,I){var i,e,r,n,g=this,t=g.s;if(I=(A=new x(A,I)).s,!t||!I)return new x(NaN);if(t!=I)return A.s=-I,g.plus(A);var o=g.e/14,C=A.e/14,a=g.c,h=A.c;if(!o||!C){if(!a||!h)return a?(A.s=-I,A):new x(h?g:NaN);if(!a[0]||!h[0])return h[0]?(A.s=-I,A):new x(a[0]?g:3==G?-0:0)}if(o=f(o),C=f(C),a=a.slice(),t=o-C){for((n=t<0)?(t=-t,r=a):(C=o,r=h),r.reverse(),I=t;I--;r.push(0));r.reverse()}else for(e=(n=(t=a.length)<(I=h.length))?t:I,t=I=0;I<e;I++)if(a[I]!=h[I]){n=a[I]<h[I];break}if(n&&(r=a,a=h,h=r,A.s=-A.s),(I=(e=h.length)-(i=a.length))>0)for(;I--;a[i++]=0);for(I=1e14-1;e>t;){if(a[--e]<h[e]){for(i=e;i&&!a[--i];a[i]=I);--a[i],a[e]+=1e14}a[e]-=h[e]}for(;0==a[0];a.splice(0,1),--C);return a[0]?L(A,a,C):(A.s=3==G?-1:1,A.c=[A.e=0],A)},y.modulo=y.mod=function(A,I){var e,r,n=this;return A=new x(A,I),!n.c||!A.s||A.c&&!A.c[0]?new x(NaN):!A.c||n.c&&!n.c[0]?new x(n):(9==m?(r=A.s,A.s=1,e=i(n,A,0,3),A.s=r,e.s*=r):e=i(n,A,0,m),(A=n.minus(e.times(A))).c[0]||1!=m||(A.s=n.s),A)},y.multipliedBy=y.times=function(A,I){var i,e,r,n,g,t,o,C,a,h,u,s,Q,B=this,E=B.c,c=(A=new x(A,I)).c;if(!(E&&c&&E[0]&&c[0]))return!B.s||!A.s||E&&!E[0]&&!c||c&&!c[0]&&!E?A.c=A.e=A.s=null:(A.s*=B.s,E&&c?(A.c=[0],A.e=0):A.c=A.e=null),A;for(e=f(B.e/14)+f(A.e/14),A.s*=B.s,(o=E.length)<(h=c.length)&&(Q=E,E=c,c=Q,r=o,o=h,h=r),r=o+h,Q=[];r--;Q.push(0));for(1e14,1e7,r=h;--r>=0;){for(i=0,u=c[r]%1e7,s=c[r]/1e7|0,n=r+(g=o);n>r;)i=((C=u*(C=E[--g]%1e7)+(t=s*C+(a=E[g]/1e7|0)*u)%1e7*1e7+Q[n]+i)/1e14|0)+(t/1e7|0)+s*a,Q[n--]=C%1e14;Q[n]=i}return i?++e:Q.splice(0,1),L(A,Q,e)},y.negated=function(){var A=new x(this);return A.s=-A.s||null,A},y.plus=function(A,I){var i,e=this,r=e.s;if(I=(A=new x(A,I)).s,!r||!I)return new x(NaN);if(r!=I)return A.s=-I,e.minus(A);var n=e.e/14,g=A.e/14,t=e.c,o=A.c;if(!n||!g){if(!t||!o)return new x(r/0);if(!t[0]||!o[0])return o[0]?A:new x(t[0]?e:0*r)}if(n=f(n),g=f(g),t=t.slice(),r=n-g){for(r>0?(g=n,i=o):(r=-r,i=t),i.reverse();r--;i.push(0));i.reverse()}for((r=t.length)-(I=o.length)<0&&(i=o,o=t,t=i,I=r),r=0;I;)r=(t[--I]=t[I]+o[I]+r)/1e14|0,t[I]=1e14===t[I]?0:t[I]%1e14;return r&&(t=[r].concat(t),++g),L(A,t,g)},y.precision=y.sd=function(A,I){var i,e,r,n=this;if(null!=A&&A!==!!A)return Q(A,1,1e9),null==I?I=G:Q(I,0,8),O(new x(n),A,I);if(!(i=n.c))return null;if(e=14*(r=i.length-1)+1,r=i[r]){for(;r%10==0;r/=10,e--);for(r=i[0];r>=10;r/=10,e++);}return A&&n.e+1>e&&(e=n.e+1),e},y.shiftedBy=function(A){return Q(A,-9007199254740991,9007199254740991),this.times("1e"+A)},y.squareRoot=y.sqrt=function(){var A,I,e,r,n,g=this,t=g.c,o=g.s,C=g.e,a=p+4,h=new x("0.5");if(1!==o||!t||!t[0])return new x(!o||o<0&&(!t||t[0])?NaN:t?g:1/0);if(0==(o=Math.sqrt(+P(g)))||o==1/0?(((I=u(t)).length+C)%2==0&&(I+="0"),o=Math.sqrt(+I),C=f((C+1)/2)-(C<0||C%2),e=new x(I=o==1/0?"1e"+C:(I=o.toExponential()).slice(0,I.indexOf("e")+1)+C)):e=new x(o+""),e.c[0])for((o=(C=e.e)+a)<3&&(o=0);;)if(n=e,e=h.times(n.plus(i(g,n,a,1))),u(n.c).slice(0,o)===(I=u(e.c)).slice(0,o)){if(e.e<C&&--o,"9999"!=(I=I.slice(o-3,o+1))&&(r||"4999"!=I)){+I&&(+I.slice(1)||"5"!=I.charAt(0))||(O(e,e.e+p+2,1),A=!e.times(e).eq(g));break}if(!r&&(O(n,n.e+p+2,0),n.times(n).eq(g))){e=n;break}a+=4,o+=4,r=1}return O(e,e.e+p+1,G,A)},y.toExponential=function(A,I){return null!=A&&(Q(A,0,1e9),A++),K(this,A,I,1)},y.toFixed=function(A,I){return null!=A&&(Q(A,0,1e9),A=A+this.e+1),K(this,A,I)},y.toFormat=function(A,I,i){var e,r=this;if(null==i)null!=A&&I&&"object"==typeof I?(i=I,I=null):A&&"object"==typeof A?(i=A,A=I=null):i=Y;else if("object"!=typeof i)throw Error(C+"Argument not an object: "+i);if(e=r.toFixed(A,I),r.c){var n,g=e.split("."),t=+i.groupSize,o=+i.secondaryGroupSize,a=i.groupSeparator||"",h=g[0],f=g[1],u=r.s<0,s=u?h.slice(1):h,Q=s.length;if(o&&(n=t,t=o,o=n,Q-=n),t>0&&Q>0){for(n=Q%t||t,h=s.substr(0,n);n<Q;n+=t)h+=a+s.substr(n,t);o>0&&(h+=a+s.slice(n)),u&&(h="-"+h)}e=f?h+(i.decimalSeparator||"")+((o=+i.fractionGroupSize)?f.replace(new RegExp("\\d{"+o+"}\\B","g"),"$&"+(i.fractionGroupSeparator||"")):f):h}return(i.prefix||"")+e+(i.suffix||"")},y.toFraction=function(A){var I,e,r,n,g,t,o,a,f,s,Q,B,E=this,c=E.c;if(null!=A&&(!(o=new x(A)).isInteger()&&(o.c||1!==o.s)||o.lt(d)))throw Error(C+"Argument "+(o.isInteger()?"out of range: ":"not an integer: ")+P(o));if(!c)return new x(E);for(I=new x(d),f=e=new x(d),r=a=new x(d),B=u(c),g=I.e=B.length-E.e-1,I.c[0]=h[(t=g%14)<0?14+t:t],A=!A||o.comparedTo(I)>0?g>0?I:f:o,t=b,b=1/0,o=new x(B),a.c[0]=0;s=i(o,I,0,1),1!=(n=e.plus(s.times(r))).comparedTo(A);)e=r,r=n,f=a.plus(s.times(n=f)),a=n,I=o.minus(s.times(n=I)),o=n;return n=i(A.minus(e),r,0,1),a=a.plus(n.times(f)),e=e.plus(n.times(r)),a.s=f.s=E.s,Q=i(f,r,g*=2,G).minus(E).abs().comparedTo(i(a,e,g,G).minus(E).abs())<1?[f,r]:[a,e],b=t,Q},y.toNumber=function(){return+P(this)},y.toPrecision=function(A,I){return null!=A&&Q(A,1,1e9),K(this,A,I,2)},y.toString=function(A){var I,i=this,r=i.s,n=i.e;return null===n?r?(I="Infinity",r<0&&(I="-"+I)):I="NaN":(null==A?I=n<=D||n>=v?E(u(i.c),n):c(u(i.c),n,"0"):10===A?I=c(u((i=O(new x(i),p+n+1,G)).c),i.e,"0"):(Q(A,2,N.length,"Base"),I=e(c(u(i.c),n,"0"),10,A,r,!0)),r<0&&i.c[0]&&(I="-"+I)),I},y.valueOf=y.toJSON=function(){return P(this)},y._isBigNumber=!0,y[Symbol.toStringTag]="BigNumber",y[Symbol.for("nodejs.util.inspect.custom")]=y.valueOf,null!=I&&x.set(I),x}();function w(A){return(4294967296+A).toString(16).substring(1)}var U={normalizeInput:function(A){var I;if(A instanceof Uint8Array)I=A;else if(A instanceof Buffer)I=new Uint8Array(A);else{if("string"!=typeof A)throw new Error("Input must be an string, Buffer or Uint8Array");I=new Uint8Array(Buffer.from(A,"utf8"))}return I},toHex:function(A){return Array.prototype.map.call(A,(function(A){return(A<16?"0":"")+A.toString(16)})).join("")},debugPrint:function(A,I,i){for(var e="\n"+A+" = ",r=0;r<I.length;r+=2){if(32===i)e+=w(I[r]).toUpperCase(),e+=" ",e+=w(I[r+1]).toUpperCase();else{if(64!==i)throw new Error("Invalid size "+i);e+=w(I[r+1]).toUpperCase(),e+=w(I[r]).toUpperCase()}r%6==4?e+="\n"+new Array(A.length+4).join(" "):r<I.length-2&&(e+=" ")}console.log(e)},testSpeed:function(A,I,i){for(var e=(new Date).getTime(),r=new Uint8Array(I),n=0;n<I;n++)r[n]=n%256;var g=(new Date).getTime();for(console.log("Generated random input in "+(g-e)+"ms"),e=g,n=0;n<i;n++){var t=A(r),o=(new Date).getTime(),C=o-e;e=o,console.log("Hashed in "+C+"ms: "+t.substring(0,20)+"..."),console.log(Math.round(I/(1<<20)/(C/1e3)*100)/100+" MB PER SECOND")}}};function S(A,I,i){var e=A[I]+A[i],r=A[I+1]+A[i+1];e>=4294967296&&r++,A[I]=e,A[I+1]=r}function F(A,I,i,e){var r=A[I]+i;i<0&&(r+=4294967296);var n=A[I+1]+e;r>=4294967296&&n++,A[I]=r,A[I+1]=n}function y(A,I){return A[I]^A[I+1]<<8^A[I+2]<<16^A[I+3]<<24}function d(A,I,i,e,r,n){var g=v[r],t=v[r+1],o=v[n],C=v[n+1];S(D,A,I),F(D,A,g,t);var a=D[e]^D[A],h=D[e+1]^D[A+1];D[e]=h,D[e+1]=a,S(D,i,e),a=D[I]^D[i],h=D[I+1]^D[i+1],D[I]=a>>>24^h<<8,D[I+1]=h>>>24^a<<8,S(D,A,I),F(D,A,o,C),a=D[e]^D[A],h=D[e+1]^D[A+1],D[e]=a>>>16^h<<16,D[e+1]=h>>>16^a<<16,S(D,i,e),a=D[I]^D[i],h=D[I+1]^D[i+1],D[I]=h>>>31^a<<1,D[I+1]=a>>>31^h<<1}var p=new Uint32Array([4089235720,1779033703,2227873595,3144134277,4271175723,1013904242,1595750129,2773480762,2917565137,1359893119,725511199,2600822924,4215389547,528734635,327033209,1541459225]),G=new Uint8Array([0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,14,10,4,8,9,15,13,6,1,12,0,2,11,7,5,3,11,8,12,0,5,2,15,13,10,14,3,6,7,1,9,4,7,9,3,1,13,12,11,14,2,6,5,10,4,0,15,8,9,0,5,7,2,4,10,15,14,1,11,12,6,8,3,13,2,12,6,10,0,11,8,3,4,13,7,5,15,14,1,9,12,5,1,15,14,13,4,10,0,7,6,3,9,2,8,11,13,11,7,14,12,1,3,9,5,0,15,4,8,6,2,10,6,15,14,9,11,3,0,8,12,2,13,7,1,4,10,5,10,2,8,4,7,6,1,5,15,11,9,14,3,12,13,0,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,14,10,4,8,9,15,13,6,1,12,0,2,11,7,5,3].map((function(A){return 2*A}))),D=new Uint32Array(32),v=new Uint32Array(32);function k(A,I){var i=0;for(i=0;i<16;i++)D[i]=A.h[i],D[i+16]=p[i];for(D[24]=D[24]^A.t,D[25]=D[25]^A.t/4294967296,I&&(D[28]=~D[28],D[29]=~D[29]),i=0;i<32;i++)v[i]=y(A.b,4*i);for(i=0;i<12;i++)d(0,8,16,24,G[16*i+0],G[16*i+1]),d(2,10,18,26,G[16*i+2],G[16*i+3]),d(4,12,20,28,G[16*i+4],G[16*i+5]),d(6,14,22,30,G[16*i+6],G[16*i+7]),d(0,10,20,30,G[16*i+8],G[16*i+9]),d(2,12,22,24,G[16*i+10],G[16*i+11]),d(4,14,16,26,G[16*i+12],G[16*i+13]),d(6,8,18,28,G[16*i+14],G[16*i+15]);for(i=0;i<16;i++)A.h[i]=A.h[i]^D[i]^D[i+16]}function b(A,I){if(0===A||A>64)throw new Error("Illegal output length, expected 0 < length <= 64");if(I&&I.length>64)throw new Error("Illegal key, expected Uint8Array with 0 < length <= 64");for(var i={b:new Uint8Array(128),h:new Uint32Array(16),t:0,c:0,outlen:A},e=0;e<16;e++)i.h[e]=p[e];var r=I?I.length:0;return i.h[0]^=16842752^r<<8^A,I&&(H(i,I),i.c=128),i}function H(A,I){for(var i=0;i<I.length;i++)128===A.c&&(A.t+=A.c,k(A,!1),A.c=0),A.b[A.c++]=I[i]}function m(A){for(A.t+=A.c;A.c<128;)A.b[A.c++]=0;k(A,!0);for(var I=new Uint8Array(A.outlen),i=0;i<A.outlen;i++)I[i]=A.h[i>>2]>>8*(3&i);return I}function M(A,I,i){i=i||64,A=U.normalizeInput(A);var e=b(i,I);return H(e,A),m(e)}var Y={blake2b:M,blake2bHex:function(A,I,i){var e=M(A,I,i);return U.toHex(e)},blake2bInit:b,blake2bUpdate:H,blake2bFinal:m};function N(A,I){return A[I]^A[I+1]<<8^A[I+2]<<16^A[I+3]<<24}function x(A,I,i,e,r,n){O[A]=O[A]+O[I]+r,O[e]=K(O[e]^O[A],16),O[i]=O[i]+O[e],O[I]=K(O[I]^O[i],12),O[A]=O[A]+O[I]+n,O[e]=K(O[e]^O[A],8),O[i]=O[i]+O[e],O[I]=K(O[I]^O[i],7)}function K(A,I){return A>>>I^A<<32-I}var R=new Uint32Array([1779033703,3144134277,1013904242,2773480762,1359893119,2600822924,528734635,1541459225]),L=new Uint8Array([0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,14,10,4,8,9,15,13,6,1,12,0,2,11,7,5,3,11,8,12,0,5,2,15,13,10,14,3,6,7,1,9,4,7,9,3,1,13,12,11,14,2,6,5,10,4,0,15,8,9,0,5,7,2,4,10,15,14,1,11,12,6,8,3,13,2,12,6,10,0,11,8,3,4,13,7,5,15,14,1,9,12,5,1,15,14,13,4,10,0,7,6,3,9,2,8,11,13,11,7,14,12,1,3,9,5,0,15,4,8,6,2,10,6,15,14,9,11,3,0,8,12,2,13,7,1,4,10,5,10,2,8,4,7,6,1,5,15,11,9,14,3,12,13,0]),O=new Uint32Array(16),P=new Uint32Array(16);function j(A,I){var i=0;for(i=0;i<8;i++)O[i]=A.h[i],O[i+8]=R[i];for(O[12]^=A.t,O[13]^=A.t/4294967296,I&&(O[14]=~O[14]),i=0;i<16;i++)P[i]=N(A.b,4*i);for(i=0;i<10;i++)x(0,4,8,12,P[L[16*i+0]],P[L[16*i+1]]),x(1,5,9,13,P[L[16*i+2]],P[L[16*i+3]]),x(2,6,10,14,P[L[16*i+4]],P[L[16*i+5]]),x(3,7,11,15,P[L[16*i+6]],P[L[16*i+7]]),x(0,5,10,15,P[L[16*i+8]],P[L[16*i+9]]),x(1,6,11,12,P[L[16*i+10]],P[L[16*i+11]]),x(2,7,8,13,P[L[16*i+12]],P[L[16*i+13]]),x(3,4,9,14,P[L[16*i+14]],P[L[16*i+15]]);for(i=0;i<8;i++)A.h[i]^=O[i]^O[i+8]}function J(A,I){if(!(A>0&&A<=32))throw new Error("Incorrect output length, should be in [1, 32]");var i=I?I.length:0;if(I&&!(i>0&&i<=32))throw new Error("Incorrect key length, should be in [1, 32]");var e={h:new Uint32Array(R),b:new Uint32Array(64),c:0,t:0,outlen:A};return e.h[0]^=16842752^i<<8^A,i>0&&(X(e,I),e.c=64),e}function X(A,I){for(var i=0;i<I.length;i++)64===A.c&&(A.t+=A.c,j(A,!1),A.c=0),A.b[A.c++]=I[i]}function T(A){for(A.t+=A.c;A.c<64;)A.b[A.c++]=0;j(A,!0);for(var I=new Uint8Array(A.outlen),i=0;i<A.outlen;i++)I[i]=A.h[i>>2]>>8*(3&i)&255;return I}function V(A,I,i){i=i||32,A=U.normalizeInput(A);var e=J(i,I);return X(e,A),T(e)}var q,Z={blake2s:V,blake2sHex:function(A,I,i){var e=V(A,I,i);return U.toHex(e)},blake2sInit:J,blake2sUpdate:X,blake2sFinal:T},W={blake2b:Y.blake2b,blake2bHex:Y.blake2bHex,blake2bInit:Y.blake2bInit,blake2bUpdate:Y.blake2bUpdate,blake2bFinal:Y.blake2bFinal,blake2s:Z.blake2s,blake2sHex:Z.blake2sHex,blake2sInit:Z.blake2sInit,blake2sUpdate:Z.blake2sUpdate,blake2sFinal:Z.blake2sFinal},_=W.blake2b,z=W.blake2bInit,$=W.blake2bUpdate,AA=W.blake2bFinal;if("[object process]"===Object.prototype.toString.call("undefined"!=typeof process?process:0)){var IA=require("util").promisify;q=IA(require("crypto").randomFill)}else q=function(A){return new Promise((function(I){crypto.getRandomValues(A),I()}))};function iA(A){if(!A)return"";for(var I="",i=0;i<A.length;i++){var e=(255&A[i]).toString(16);I+=e=1===e.length?"0"+e:e}return I.toUpperCase()}function eA(A){if(!A)return new Uint8Array;for(var I=[],i=0;i<A.length;i+=2)I.push(parseInt(A.substr(i,2),16));return new Uint8Array(I)}var rA="13456789abcdefghijkmnopqrstuwxyz";function nA(A){for(var I=A.length,i=8*I%5,e=0===i?0:5-i,r=0,n="",g=0,t=0;t<I;t++)for(r=r<<8|A[t],g+=8;g>=5;)n+=rA[r>>>g+e-5&31],g-=5;return g>0&&(n+=rA[r<<5-(g+e)&31]),n}function gA(A){var I=rA.indexOf(A);if(-1===I)throw new Error("Invalid character found: "+A);return I}function tA(A){for(var I=A.length,i=5*I%8,e=0===i?0:8-i,r=0,n=0,g=0,t=new Uint8Array(Math.ceil(5*I/8)),o=0;o<I;o++)n=n<<5|gA(A[o]),(r+=5)>=8&&(t[g++]=n>>>r+e-8&255,r-=8);return r>0&&(t[g++]=n<<r+e-8&255),0!==i&&(t=t.slice(1)),t}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */function oA(A){var I,i={valid:!1,publicKeyBytes:null};if(!hA(A)||!/^(xrb_|nano_)[13][13-9a-km-uw-z]{59}$/.test(A))return i;I=A.startsWith("xrb_")?4:5;var e=tA(A.substr(I,52));return function(A,I){for(var i=0;i<A.length;i++)if(A[i]!==I[i])return!1;return!0}(tA(A.substr(I+52)),_(e,null,5).reverse())?{publicKeyBytes:e,valid:!0}:i}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */var CA=Math.pow(2,32)-1,aA=new l("0xffffffffffffffffffffffffffffffff");function hA(A){return"string"==typeof A}function fA(A){return"0"===A||!(!hA(A)||!/^[1-9]{1}[0-9]{0,38}$/.test(A))&&new l(A).isLessThanOrEqualTo(aA)}function uA(A){return hA(A)&&/^[0-9a-fA-F]{64}$/.test(A)}function sA(A){return hA(A)&&/^[0-9a-fA-F]{16}$/.test(A)}function QA(A){return Number.isInteger(A)&&A>=0&&A<=CA}function BA(A){return uA(A)}function EA(A){return uA(A)}function cA(A){return oA(A).valid}function lA(A){return hA(A)&&/^[0-9a-fA-F]{16}$/.test(A)}function wA(A){return hA(A)&&/^[0-9a-fA-F]{128}$/.test(A)}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */var UA={loaded:!1,work:null};function SA(){return new Promise((function(A,I){if(UA.loaded)return A(UA);try{n().then((function(I){var i=Object.assign(UA,{loaded:!0,work:I.cwrap("emscripten_work","string",["string","string","number","number"])});A(i)}))}catch(A){I(A)}}))}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */
var FA=function(A){var I=new Float64Array(16);if(A)for(var i=0;i<A.length;i++)I[i]=A[i];return I};new Uint8Array(32)[0]=9;var yA=FA(),dA=FA([1]),pA=FA([30883,4953,19914,30187,55467,16705,2637,112,59544,30585,16505,36039,65139,11119,27886,20995]),GA=FA([61785,9906,39828,60374,45398,33411,5274,224,53552,61171,33010,6542,64743,22239,55772,9222]),DA=FA([54554,36645,11616,51542,42930,38181,51040,26924,56412,64982,57905,49316,21502,52590,14035,8553]),vA=FA([26200,26214,26214,26214,26214,26214,26214,26214,26214,26214,26214,26214,26214,26214,26214,26214]),kA=FA([41136,18958,6951,50414,58488,44335,6150,12099,55207,15867,153,11085,57099,20417,9344,11139]);function bA(A,I,i,e){return function(A,I,i,e,r){for(var n=0,g=0;g<r;g++)n|=A[I+g]^i[e+g];return(1&n-1>>>8)-1}(A,I,i,e,32)}function HA(A,I){var i;for(i=0;i<16;i++)A[i]=0|I[i]}function mA(A){for(var I,i=1,e=0;e<16;e++)I=A[e]+i+65535,i=Math.floor(I/65536),A[e]=I-65536*i;A[0]+=i-1+37*(i-1)}function MA(A,I,i){for(var e,r=~(i-1),n=0;n<16;n++)e=r&(A[n]^I[n]),A[n]^=e,I[n]^=e}function YA(A,I){for(var i,e=FA(),r=FA(),n=0;n<16;n++)r[n]=I[n];mA(r),mA(r),mA(r);for(var g=0;g<2;g++){e[0]=r[0]-65517;for(n=1;n<15;n++)e[n]=r[n]-65535-(e[n-1]>>16&1),e[n-1]&=65535;e[15]=r[15]-32767-(e[14]>>16&1),i=e[15]>>16&1,e[14]&=65535,MA(r,e,1-i)}for(n=0;n<16;n++)A[2*n]=255&r[n],A[2*n+1]=r[n]>>8}function NA(A,I){var i=new Uint8Array(32),e=new Uint8Array(32);return YA(i,A),YA(e,I),bA(i,0,e,0)}function xA(A){var I=new Uint8Array(32);return YA(I,A),1&I[0]}function KA(A,I,i){for(var e=0;e<16;e++)A[e]=I[e]+i[e]}function RA(A,I,i){for(var e=0;e<16;e++)A[e]=I[e]-i[e]}function LA(A,I,i){var e,r,n=0,g=0,t=0,o=0,C=0,a=0,h=0,f=0,u=0,s=0,Q=0,B=0,E=0,c=0,l=0,w=0,U=0,S=0,F=0,y=0,d=0,p=0,G=0,D=0,v=0,k=0,b=0,H=0,m=0,M=0,Y=0,N=i[0],x=i[1],K=i[2],R=i[3],L=i[4],O=i[5],P=i[6],j=i[7],J=i[8],X=i[9],T=i[10],V=i[11],q=i[12],Z=i[13],W=i[14],_=i[15];n+=(e=I[0])*N,g+=e*x,t+=e*K,o+=e*R,C+=e*L,a+=e*O,h+=e*P,f+=e*j,u+=e*J,s+=e*X,Q+=e*T,B+=e*V,E+=e*q,c+=e*Z,l+=e*W,w+=e*_,g+=(e=I[1])*N,t+=e*x,o+=e*K,C+=e*R,a+=e*L,h+=e*O,f+=e*P,u+=e*j,s+=e*J,Q+=e*X,B+=e*T,E+=e*V,c+=e*q,l+=e*Z,w+=e*W,U+=e*_,t+=(e=I[2])*N,o+=e*x,C+=e*K,a+=e*R,h+=e*L,f+=e*O,u+=e*P,s+=e*j,Q+=e*J,B+=e*X,E+=e*T,c+=e*V,l+=e*q,w+=e*Z,U+=e*W,S+=e*_,o+=(e=I[3])*N,C+=e*x,a+=e*K,h+=e*R,f+=e*L,u+=e*O,s+=e*P,Q+=e*j,B+=e*J,E+=e*X,c+=e*T,l+=e*V,w+=e*q,U+=e*Z,S+=e*W,F+=e*_,C+=(e=I[4])*N,a+=e*x,h+=e*K,f+=e*R,u+=e*L,s+=e*O,Q+=e*P,B+=e*j,E+=e*J,c+=e*X,l+=e*T,w+=e*V,U+=e*q,S+=e*Z,F+=e*W,y+=e*_,a+=(e=I[5])*N,h+=e*x,f+=e*K,u+=e*R,s+=e*L,Q+=e*O,B+=e*P,E+=e*j,c+=e*J,l+=e*X,w+=e*T,U+=e*V,S+=e*q,F+=e*Z,y+=e*W,d+=e*_,h+=(e=I[6])*N,f+=e*x,u+=e*K,s+=e*R,Q+=e*L,B+=e*O,E+=e*P,c+=e*j,l+=e*J,w+=e*X,U+=e*T,S+=e*V,F+=e*q,y+=e*Z,d+=e*W,p+=e*_,f+=(e=I[7])*N,u+=e*x,s+=e*K,Q+=e*R,B+=e*L,E+=e*O,c+=e*P,l+=e*j,w+=e*J,U+=e*X,S+=e*T,F+=e*V,y+=e*q,d+=e*Z,p+=e*W,G+=e*_,u+=(e=I[8])*N,s+=e*x,Q+=e*K,B+=e*R,E+=e*L,c+=e*O,l+=e*P,w+=e*j,U+=e*J,S+=e*X,F+=e*T,y+=e*V,d+=e*q,p+=e*Z,G+=e*W,D+=e*_,s+=(e=I[9])*N,Q+=e*x,B+=e*K,E+=e*R,c+=e*L,l+=e*O,w+=e*P,U+=e*j,S+=e*J,F+=e*X,y+=e*T,d+=e*V,p+=e*q,G+=e*Z,D+=e*W,v+=e*_,Q+=(e=I[10])*N,B+=e*x,E+=e*K,c+=e*R,l+=e*L,w+=e*O,U+=e*P,S+=e*j,F+=e*J,y+=e*X,d+=e*T,p+=e*V,G+=e*q,D+=e*Z,v+=e*W,k+=e*_,B+=(e=I[11])*N,E+=e*x,c+=e*K,l+=e*R,w+=e*L,U+=e*O,S+=e*P,F+=e*j,y+=e*J,d+=e*X,p+=e*T,G+=e*V,D+=e*q,v+=e*Z,k+=e*W,b+=e*_,E+=(e=I[12])*N,c+=e*x,l+=e*K,w+=e*R,U+=e*L,S+=e*O,F+=e*P,y+=e*j,d+=e*J,p+=e*X,G+=e*T,D+=e*V,v+=e*q,k+=e*Z,b+=e*W,H+=e*_,c+=(e=I[13])*N,l+=e*x,w+=e*K,U+=e*R,S+=e*L,F+=e*O,y+=e*P,d+=e*j,p+=e*J,G+=e*X,D+=e*T,v+=e*V,k+=e*q,b+=e*Z,H+=e*W,m+=e*_,l+=(e=I[14])*N,w+=e*x,U+=e*K,S+=e*R,F+=e*L,y+=e*O,d+=e*P,p+=e*j,G+=e*J,D+=e*X,v+=e*T,k+=e*V,b+=e*q,H+=e*Z,m+=e*W,M+=e*_,w+=(e=I[15])*N,g+=38*(S+=e*K),t+=38*(F+=e*R),o+=38*(y+=e*L),C+=38*(d+=e*O),a+=38*(p+=e*P),h+=38*(G+=e*j),f+=38*(D+=e*J),u+=38*(v+=e*X),s+=38*(k+=e*T),Q+=38*(b+=e*V),B+=38*(H+=e*q),E+=38*(m+=e*Z),c+=38*(M+=e*W),l+=38*(Y+=e*_),n=(e=(n+=38*(U+=e*x))+(r=1)+65535)-65536*(r=Math.floor(e/65536)),g=(e=g+r+65535)-65536*(r=Math.floor(e/65536)),t=(e=t+r+65535)-65536*(r=Math.floor(e/65536)),o=(e=o+r+65535)-65536*(r=Math.floor(e/65536)),C=(e=C+r+65535)-65536*(r=Math.floor(e/65536)),a=(e=a+r+65535)-65536*(r=Math.floor(e/65536)),h=(e=h+r+65535)-65536*(r=Math.floor(e/65536)),f=(e=f+r+65535)-65536*(r=Math.floor(e/65536)),u=(e=u+r+65535)-65536*(r=Math.floor(e/65536)),s=(e=s+r+65535)-65536*(r=Math.floor(e/65536)),Q=(e=Q+r+65535)-65536*(r=Math.floor(e/65536)),B=(e=B+r+65535)-65536*(r=Math.floor(e/65536)),E=(e=E+r+65535)-65536*(r=Math.floor(e/65536)),c=(e=c+r+65535)-65536*(r=Math.floor(e/65536)),l=(e=l+r+65535)-65536*(r=Math.floor(e/65536)),w=(e=w+r+65535)-65536*(r=Math.floor(e/65536)),n=(e=(n+=r-1+37*(r-1))+(r=1)+65535)-65536*(r=Math.floor(e/65536)),g=(e=g+r+65535)-65536*(r=Math.floor(e/65536)),t=(e=t+r+65535)-65536*(r=Math.floor(e/65536)),o=(e=o+r+65535)-65536*(r=Math.floor(e/65536)),C=(e=C+r+65535)-65536*(r=Math.floor(e/65536)),a=(e=a+r+65535)-65536*(r=Math.floor(e/65536)),h=(e=h+r+65535)-65536*(r=Math.floor(e/65536)),f=(e=f+r+65535)-65536*(r=Math.floor(e/65536)),u=(e=u+r+65535)-65536*(r=Math.floor(e/65536)),s=(e=s+r+65535)-65536*(r=Math.floor(e/65536)),Q=(e=Q+r+65535)-65536*(r=Math.floor(e/65536)),B=(e=B+r+65535)-65536*(r=Math.floor(e/65536)),E=(e=E+r+65535)-65536*(r=Math.floor(e/65536)),c=(e=c+r+65535)-65536*(r=Math.floor(e/65536)),l=(e=l+r+65535)-65536*(r=Math.floor(e/65536)),w=(e=w+r+65535)-65536*(r=Math.floor(e/65536)),n+=r-1+37*(r-1),A[0]=n,A[1]=g,A[2]=t,A[3]=o,A[4]=C,A[5]=a,A[6]=h,A[7]=f,A[8]=u,A[9]=s,A[10]=Q,A[11]=B,A[12]=E,A[13]=c,A[14]=l,A[15]=w}function OA(A,I){LA(A,I,I)}function PA(A,I,i){for(var e=new Uint8Array(i),r=0;r<i;++r)e[r]=I[r];var n=W.blake2b(e);for(r=0;r<64;++r)A[r]=n[r];return 0}function jA(A,I){var i=FA(),e=FA(),r=FA(),n=FA(),g=FA(),t=FA(),o=FA(),C=FA(),a=FA();RA(i,A[1],A[0]),RA(a,I[1],I[0]),LA(i,i,a),KA(e,A[0],A[1]),KA(a,I[0],I[1]),LA(e,e,a),LA(r,A[3],I[3]),LA(r,r,GA),LA(n,A[2],I[2]),KA(n,n,n),RA(g,e,i),RA(t,n,r),KA(o,n,r),KA(C,e,i),LA(A[0],g,t),LA(A[1],C,o),LA(A[2],o,t),LA(A[3],g,C)}function JA(A,I,i){var e;for(e=0;e<4;e++)MA(A[e],I[e],i)}function XA(A,I){var i=FA(),e=FA(),r=FA();!function(A,I){var i,e=FA();for(i=0;i<16;i++)e[i]=I[i];for(i=253;i>=0;i--)OA(e,e),2!==i&&4!==i&&LA(e,e,I);for(i=0;i<16;i++)A[i]=e[i]}(r,I[2]),LA(i,I[0],r),LA(e,I[1],r),YA(A,e),A[31]^=xA(i)<<7}function TA(A,I,i){var e,r;for(HA(A[0],yA),HA(A[1],dA),HA(A[2],dA),HA(A[3],yA),r=255;r>=0;--r)JA(A,I,e=i[r/8|0]>>(7&r)&1),jA(I,A),jA(A,A),JA(A,I,e)}function VA(A,I){var i=[FA(),FA(),FA(),FA()];HA(i[0],DA),HA(i[1],vA),HA(i[2],dA),LA(i[3],DA,vA),TA(A,i,I)}var qA,ZA=new Float64Array([237,211,245,92,26,99,18,88,214,156,247,162,222,249,222,20,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,16]);function WA(A,I){var i,e,r,n;for(e=63;e>=32;--e){for(i=0,r=e-32,n=e-12;r<n;++r)I[r]+=i-16*I[e]*ZA[r-(e-32)],i=I[r]+128>>8,I[r]-=256*i;I[r]+=i,I[e]=0}for(i=0,r=0;r<32;r++)I[r]+=i-(I[31]>>4)*ZA[r],i=I[r]>>8,I[r]&=255;for(r=0;r<32;r++)I[r]-=i*ZA[r];for(e=0;e<32;e++)I[e+1]+=I[e]>>8,A[e]=255&I[e]}function _A(A){for(var I=new Float64Array(64),i=0;i<64;i++)I[i]=A[i];for(i=0;i<64;i++)A[i]=0;WA(A,I)}function zA(A){var I=new Uint8Array(64),i=[FA(),FA(),FA(),FA()],e=new Uint8Array(32),r=W.blake2bInit(64);return W.blake2bUpdate(r,A),(I=W.blake2bFinal(r))[0]&=248,I[31]&=127,I[31]|=64,VA(i,I),XA(e,i),e}function $A(A,I){var i=FA(),e=FA(),r=FA(),n=FA(),g=FA(),t=FA(),o=FA();return HA(A[2],dA),function(A,I){var i;for(i=0;i<16;i++)A[i]=I[2*i]+(I[2*i+1]<<8);A[15]&=32767}(A[1],I),OA(r,A[1]),LA(n,r,pA),RA(r,r,A[2]),KA(n,A[2],n),OA(g,n),OA(t,g),LA(o,t,g),LA(i,o,r),LA(i,i,n),function(A,I){var i,e=FA();for(i=0;i<16;i++)e[i]=I[i];for(i=250;i>=0;i--)OA(e,e),1!==i&&LA(e,e,I);for(i=0;i<16;i++)A[i]=e[i]}(i,i),LA(i,i,r),LA(i,i,n),LA(i,i,n),LA(A[0],i,n),OA(e,A[0]),LA(e,e,n),NA(e,r)&&LA(A[0],A[0],kA),OA(e,A[0]),LA(e,e,n),NA(e,r)?-1:(xA(A[0])===I[31]>>7&&RA(A[0],yA,A[0]),LA(A[3],A[0],A[1]),0)}function AI(A,I){if(32!==I.length)throw new Error("bad secret key size");var i=new Uint8Array(64+A.length);return function(A,I,i,e){var r,n,g=new Uint8Array(64),t=new Uint8Array(64),o=new Uint8Array(64),C=new Float64Array(64),a=[FA(),FA(),FA(),FA()],h=zA(e);PA(g,e,32),g[0]&=248,g[31]&=127,g[31]|=64;var f=i+64;for(r=0;r<i;r++)A[64+r]=I[r];for(r=0;r<32;r++)A[32+r]=g[32+r];for(PA(o,A.subarray(32),i+32),_A(o),VA(a,o),XA(A,a),r=32;r<64;r++)A[r]=h[r-32];for(PA(t,A,i+64),_A(t),r=0;r<64;r++)C[r]=0;for(r=0;r<32;r++)C[r]=o[r];for(r=0;r<32;r++)for(n=0;n<32;n++)C[r+n]+=t[r]*g[n];WA(A.subarray(32),C)}(i,A,A.length,I),i}function II(A,I,i){if(64!==I.length)throw new Error("bad signature size");if(32!==i.length)throw new Error("bad public key size");var e,r=new Uint8Array(64+A.length),n=new Uint8Array(64+A.length);for(e=0;e<64;e++)r[e]=I[e];for(e=0;e<A.length;e++)r[e+64]=A[e];return function(A,I,i,e){var r,n=new Uint8Array(32),g=new Uint8Array(64),t=[FA(),FA(),FA(),FA()],o=[FA(),FA(),FA(),FA()];if(-1,i<64)return-1;if($A(o,e))return-1;for(r=0;r<i;r++)A[r]=I[r];for(r=0;r<32;r++)A[r+32]=e[r];if(PA(g,A,i),_A(g),TA(t,o,g),VA(o,I.subarray(32)),jA(t,o),XA(n,t),i-=64,bA(I,0,n,0)){for(r=0;r<i;r++)A[r]=0;return-1}for(r=0;r<i;r++)A[r]=I[r+64];return i}(n,r,r.length,i)>=0}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */function iI(A){var I,i=EA(A),e=oA(A),r=e.valid;if(!i&&!r)throw new Error("Secret key or address is not valid");i?I=zA(eA(A)):I=e.publicKeyBytes;return iA(I)}function eI(A,I){if(void 0===I&&(I={}),!EA(A))throw new Error("Public key is not valid");var i=eA(A),e=eA(A),r="xrb_";return!0===I.useNanoPrefix&&(r="nano_"),r+nA(e)+nA(_(i,null,5).reverse())}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */(qA=A.Unit||(A.Unit={})).hex="hex",qA.raw="raw",qA.nano="nano",qA.knano="knano",qA.Nano="Nano",qA.NANO="NANO",qA.KNano="KNano",qA.MNano="MNano";var rI={hex:0,raw:0,nano:24,knano:27,Nano:30,NANO:30,KNano:33,MNano:36},nI=l.clone({EXPONENTIAL_AT:1e9,DECIMAL_PLACES:rI.MNano});function gI(A,I){var i=new Error("From or to is not valid");if(!I)throw i;var e=rI[I.from],r=rI[I.to];if(void 0===e||void 0===r)throw new Error("From or to is not valid");var n=new Error("Value is not valid");if("hex"===I.from){if(!/^[0-9a-fA-F]{32}$/.test(A))throw n}else if(!function(A){if(!hA(A))return!1;if(A.startsWith(".")||A.endsWith("."))return!1;var I=A.replace(".","");if(A.length-I.length>1)return!1;for(var i=0,e=I;i<e.length;i++){var r=e[i];if(r<"0"||r>"9")return!1}return!0}(A))throw n;var g,t=e-r;if(g="hex"===I.from?new nI("0x"+A):new nI(A),t<0)for(var o=0;o<-t;o++)g=g.dividedBy(10);else if(t>0)for(o=0;o<t;o++)g=g.multipliedBy(10);return"hex"===I.to?g.toString(16).padStart(32,"0"):g.toString()}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */var tI=new Uint8Array(32);function oI(I){var i,e=eA(iI(I.account)),r=eA(I.previous),n=eA(iI(I.representative)),g=eA(gI(I.balance,{from:A.Unit.raw,to:A.Unit.hex}));i=cA(I.link)?eA(iI(I.link)):eA(I.link);var t=z(32);return $(t,tI),$(t,e),$(t,r),$(t,n),$(t,g),$(t,i),iA(AA(t))}
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */
function CI(A){if(!BA(A.hash))throw new Error("Hash is not valid");if(!EA(A.secretKey))throw new Error("Secret key is not valid");return iA(function(A,I){for(var i=AI(A,I),e=new Uint8Array(64),r=0;r<e.length;r++)e[r]=i[r];return e}(eA(A.hash),eA(A.secretKey)))}tI[31]=6;
/*!
     * nanocurrency-js: A toolkit for the Nano cryptocurrency.
     * Copyright (c) 2019 Marvin ROGER <dev at marvinroger dot fr>
     * Licensed under GPL-3.0 (https://git.io/vAZsK)
     */
var aI="0000000000000000000000000000000000000000000000000000000000000000";A.checkAddress=cA,A.checkAmount=fA,A.checkHash=BA,A.checkIndex=QA,A.checkKey=EA,A.checkSeed=uA,A.checkSignature=wA,A.checkThreshold=sA,A.checkWork=lA,A.computeWork=function(A,I){return void 0===I&&(I={}),e(this,void 0,void 0,(function(){var i,e,n,g,t,o,C,a;return r(this,(function(r){switch(r.label){case 0:return i=I.workerIndex,e=void 0===i?0:i,n=I.workerCount,g=void 0===n?1:n,t=I.workThreshold,o=void 0===t?"ffffffc000000000":t,[4,SA()];case 1:if(C=r.sent(),!BA(A))throw new Error("Hash is not valid");if(!sA(o))throw new Error("Threshold is not valid");if(!Number.isInteger(e)||!Number.isInteger(g)||e<0||g<1||e>g-1)throw new Error("Worker parameters are not valid");return a=C.work(A,o,e,g),"1"===a[1]?[2,a.substr(2)]:[2,null]}}))}))},A.convert=gI,A.createBlock=function(A,I){if(!EA(A))throw new Error("Secret key is not valid");if(void 0===I.work)throw new Error("Work is not set");if(!cA(I.representative))throw new Error("Representative is not valid");if(!fA(I.balance))throw new Error("Balance is not valid");var i;if(null===I.previous)i=aI;else if(!BA(i=I.previous))throw new Error("Previous is not valid");var e,r=!1;if(null===I.link)e=aI;else if(cA(e=I.link))r=!0;else if(!BA(e))throw new Error("Link is not valid");if(i===aI&&(r||e===aI))throw new Error("Block is impossible");var n,g,t=eI(iI(A)),o=oI({account:t,previous:i,representative:I.representative,balance:I.balance,link:e}),C=CI({hash:o,secretKey:A});return r?n=iI(g=e):g=eI(n=e),{hash:o,block:{type:"state",account:t,previous:i,representative:I.representative,balance:I.balance,link:n,link_as_account:g,work:I.work,signature:C}}},A.deriveAddress=eI,A.derivePublicKey=iI,A.deriveSecretKey=function(A,I){if(!uA(A))throw new Error("Seed is not valid");if(!QA(I))throw new Error("Index is not valid");var i=eA(A),e=new ArrayBuffer(4);new DataView(e).setUint32(0,I);var r=new Uint8Array(e),n=z(32);return $(n,i),$(n,r),iA(AA(n))},A.generateSeed=function(){return new Promise((function(A,I){var i;(i=32,new Promise((function(A,I){var e=new Uint8Array(i);q(e).then((function(){return A(e)})).catch(I)}))).then((function(I){var i=I.reduce((function(A,I){return""+A+("0"+I.toString(16)).slice(-2)}),"");return A(i)})).catch(I)}))},A.hashBlock=function(A){if(!cA(A.account))throw new Error("Account is not valid");if(!BA(A.previous))throw new Error("Previous is not valid");if(!cA(A.representative))throw new Error("Representative is not valid");if(!fA(A.balance))throw new Error("Balance is not valid");if(!cA(A.link)&&!BA(A.link))throw new Error("Link is not valid");return oI(A)},A.signBlock=CI,A.validateWork=function(A){var I,i=null!==(I=A.threshold)&&void 0!==I?I:"ffffffc000000000";if(!BA(A.blockHash))throw new Error("Hash is not valid");if(!lA(A.work))throw new Error("Work is not valid");if(!sA(i))throw new Error("Threshold is not valid");var e=new l("0x"+i),r=eA(A.blockHash),n=eA(A.work).reverse(),g=z(8);$(g,n),$(g,r);var t=iA(AA(g).reverse());return new l("0x"+t).isGreaterThanOrEqualTo(e)},A.verifyBlock=function(A){if(!BA(A.hash))throw new Error("Hash is not valid");if(!wA(A.signature))throw new Error("Signature is not valid");if(!EA(A.publicKey))throw new Error("Public key is not valid");return II(eA(A.hash),eA(A.signature),eA(A.publicKey))},Object.defineProperty(A,"__esModule",{value:!0})}));

}).call(this)}).call(this,require('_process'),require("buffer").Buffer,"/node_modules/nanocurrency/dist")
},{"_process":6,"buffer":4,"crypto":3,"fs":3,"path":3,"util":3}],2:[function(require,module,exports){
'use strict'

exports.byteLength = byteLength
exports.toByteArray = toByteArray
exports.fromByteArray = fromByteArray

var lookup = []
var revLookup = []
var Arr = typeof Uint8Array !== 'undefined' ? Uint8Array : Array

var code = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
for (var i = 0, len = code.length; i < len; ++i) {
  lookup[i] = code[i]
  revLookup[code.charCodeAt(i)] = i
}

// Support decoding URL-safe base64 strings, as Node.js does.
// See: https://en.wikipedia.org/wiki/Base64#URL_applications
revLookup['-'.charCodeAt(0)] = 62
revLookup['_'.charCodeAt(0)] = 63

function getLens (b64) {
  var len = b64.length

  if (len % 4 > 0) {
    throw new Error('Invalid string. Length must be a multiple of 4')
  }

  // Trim off extra bytes after placeholder bytes are found
  // See: https://github.com/beatgammit/base64-js/issues/42
  var validLen = b64.indexOf('=')
  if (validLen === -1) validLen = len

  var placeHoldersLen = validLen === len
    ? 0
    : 4 - (validLen % 4)

  return [validLen, placeHoldersLen]
}

// base64 is 4/3 + up to two characters of the original data
function byteLength (b64) {
  var lens = getLens(b64)
  var validLen = lens[0]
  var placeHoldersLen = lens[1]
  return ((validLen + placeHoldersLen) * 3 / 4) - placeHoldersLen
}

function _byteLength (b64, validLen, placeHoldersLen) {
  return ((validLen + placeHoldersLen) * 3 / 4) - placeHoldersLen
}

function toByteArray (b64) {
  var tmp
  var lens = getLens(b64)
  var validLen = lens[0]
  var placeHoldersLen = lens[1]

  var arr = new Arr(_byteLength(b64, validLen, placeHoldersLen))

  var curByte = 0

  // if there are placeholders, only get up to the last complete 4 chars
  var len = placeHoldersLen > 0
    ? validLen - 4
    : validLen

  var i
  for (i = 0; i < len; i += 4) {
    tmp =
      (revLookup[b64.charCodeAt(i)] << 18) |
      (revLookup[b64.charCodeAt(i + 1)] << 12) |
      (revLookup[b64.charCodeAt(i + 2)] << 6) |
      revLookup[b64.charCodeAt(i + 3)]
    arr[curByte++] = (tmp >> 16) & 0xFF
    arr[curByte++] = (tmp >> 8) & 0xFF
    arr[curByte++] = tmp & 0xFF
  }

  if (placeHoldersLen === 2) {
    tmp =
      (revLookup[b64.charCodeAt(i)] << 2) |
      (revLookup[b64.charCodeAt(i + 1)] >> 4)
    arr[curByte++] = tmp & 0xFF
  }

  if (placeHoldersLen === 1) {
    tmp =
      (revLookup[b64.charCodeAt(i)] << 10) |
      (revLookup[b64.charCodeAt(i + 1)] << 4) |
      (revLookup[b64.charCodeAt(i + 2)] >> 2)
    arr[curByte++] = (tmp >> 8) & 0xFF
    arr[curByte++] = tmp & 0xFF
  }

  return arr
}

function tripletToBase64 (num) {
  return lookup[num >> 18 & 0x3F] +
    lookup[num >> 12 & 0x3F] +
    lookup[num >> 6 & 0x3F] +
    lookup[num & 0x3F]
}

function encodeChunk (uint8, start, end) {
  var tmp
  var output = []
  for (var i = start; i < end; i += 3) {
    tmp =
      ((uint8[i] << 16) & 0xFF0000) +
      ((uint8[i + 1] << 8) & 0xFF00) +
      (uint8[i + 2] & 0xFF)
    output.push(tripletToBase64(tmp))
  }
  return output.join('')
}

function fromByteArray (uint8) {
  var tmp
  var len = uint8.length
  var extraBytes = len % 3 // if we have 1 byte left, pad 2 bytes
  var parts = []
  var maxChunkLength = 16383 // must be multiple of 3

  // go through the array every three bytes, we'll deal with trailing stuff later
  for (var i = 0, len2 = len - extraBytes; i < len2; i += maxChunkLength) {
    parts.push(encodeChunk(uint8, i, (i + maxChunkLength) > len2 ? len2 : (i + maxChunkLength)))
  }

  // pad the end with zeros, but make sure to not forget the extra bytes
  if (extraBytes === 1) {
    tmp = uint8[len - 1]
    parts.push(
      lookup[tmp >> 2] +
      lookup[(tmp << 4) & 0x3F] +
      '=='
    )
  } else if (extraBytes === 2) {
    tmp = (uint8[len - 2] << 8) + uint8[len - 1]
    parts.push(
      lookup[tmp >> 10] +
      lookup[(tmp >> 4) & 0x3F] +
      lookup[(tmp << 2) & 0x3F] +
      '='
    )
  }

  return parts.join('')
}

},{}],3:[function(require,module,exports){

},{}],4:[function(require,module,exports){
(function (Buffer){(function (){
/*!
 * The buffer module from node.js, for the browser.
 *
 * @author   Feross Aboukhadijeh <https://feross.org>
 * @license  MIT
 */
/* eslint-disable no-proto */

'use strict'

var base64 = require('base64-js')
var ieee754 = require('ieee754')

exports.Buffer = Buffer
exports.SlowBuffer = SlowBuffer
exports.INSPECT_MAX_BYTES = 50

var K_MAX_LENGTH = 0x7fffffff
exports.kMaxLength = K_MAX_LENGTH

/**
 * If `Buffer.TYPED_ARRAY_SUPPORT`:
 *   === true    Use Uint8Array implementation (fastest)
 *   === false   Print warning and recommend using `buffer` v4.x which has an Object
 *               implementation (most compatible, even IE6)
 *
 * Browsers that support typed arrays are IE 10+, Firefox 4+, Chrome 7+, Safari 5.1+,
 * Opera 11.6+, iOS 4.2+.
 *
 * We report that the browser does not support typed arrays if the are not subclassable
 * using __proto__. Firefox 4-29 lacks support for adding new properties to `Uint8Array`
 * (See: https://bugzilla.mozilla.org/show_bug.cgi?id=695438). IE 10 lacks support
 * for __proto__ and has a buggy typed array implementation.
 */
Buffer.TYPED_ARRAY_SUPPORT = typedArraySupport()

if (!Buffer.TYPED_ARRAY_SUPPORT && typeof console !== 'undefined' &&
    typeof console.error === 'function') {
  console.error(
    'This browser lacks typed array (Uint8Array) support which is required by ' +
    '`buffer` v5.x. Use `buffer` v4.x if you require old browser support.'
  )
}

function typedArraySupport () {
  // Can typed array instances can be augmented?
  try {
    var arr = new Uint8Array(1)
    arr.__proto__ = { __proto__: Uint8Array.prototype, foo: function () { return 42 } }
    return arr.foo() === 42
  } catch (e) {
    return false
  }
}

Object.defineProperty(Buffer.prototype, 'parent', {
  enumerable: true,
  get: function () {
    if (!Buffer.isBuffer(this)) return undefined
    return this.buffer
  }
})

Object.defineProperty(Buffer.prototype, 'offset', {
  enumerable: true,
  get: function () {
    if (!Buffer.isBuffer(this)) return undefined
    return this.byteOffset
  }
})

function createBuffer (length) {
  if (length > K_MAX_LENGTH) {
    throw new RangeError('The value "' + length + '" is invalid for option "size"')
  }
  // Return an augmented `Uint8Array` instance
  var buf = new Uint8Array(length)
  buf.__proto__ = Buffer.prototype
  return buf
}

/**
 * The Buffer constructor returns instances of `Uint8Array` that have their
 * prototype changed to `Buffer.prototype`. Furthermore, `Buffer` is a subclass of
 * `Uint8Array`, so the returned instances will have all the node `Buffer` methods
 * and the `Uint8Array` methods. Square bracket notation works as expected -- it
 * returns a single octet.
 *
 * The `Uint8Array` prototype remains unmodified.
 */

function Buffer (arg, encodingOrOffset, length) {
  // Common case.
  if (typeof arg === 'number') {
    if (typeof encodingOrOffset === 'string') {
      throw new TypeError(
        'The "string" argument must be of type string. Received type number'
      )
    }
    return allocUnsafe(arg)
  }
  return from(arg, encodingOrOffset, length)
}

// Fix subarray() in ES2016. See: https://github.com/feross/buffer/pull/97
if (typeof Symbol !== 'undefined' && Symbol.species != null &&
    Buffer[Symbol.species] === Buffer) {
  Object.defineProperty(Buffer, Symbol.species, {
    value: null,
    configurable: true,
    enumerable: false,
    writable: false
  })
}

Buffer.poolSize = 8192 // not used by this implementation

function from (value, encodingOrOffset, length) {
  if (typeof value === 'string') {
    return fromString(value, encodingOrOffset)
  }

  if (ArrayBuffer.isView(value)) {
    return fromArrayLike(value)
  }

  if (value == null) {
    throw TypeError(
      'The first argument must be one of type string, Buffer, ArrayBuffer, Array, ' +
      'or Array-like Object. Received type ' + (typeof value)
    )
  }

  if (isInstance(value, ArrayBuffer) ||
      (value && isInstance(value.buffer, ArrayBuffer))) {
    return fromArrayBuffer(value, encodingOrOffset, length)
  }

  if (typeof value === 'number') {
    throw new TypeError(
      'The "value" argument must not be of type number. Received type number'
    )
  }

  var valueOf = value.valueOf && value.valueOf()
  if (valueOf != null && valueOf !== value) {
    return Buffer.from(valueOf, encodingOrOffset, length)
  }

  var b = fromObject(value)
  if (b) return b

  if (typeof Symbol !== 'undefined' && Symbol.toPrimitive != null &&
      typeof value[Symbol.toPrimitive] === 'function') {
    return Buffer.from(
      value[Symbol.toPrimitive]('string'), encodingOrOffset, length
    )
  }

  throw new TypeError(
    'The first argument must be one of type string, Buffer, ArrayBuffer, Array, ' +
    'or Array-like Object. Received type ' + (typeof value)
  )
}

/**
 * Functionally equivalent to Buffer(arg, encoding) but throws a TypeError
 * if value is a number.
 * Buffer.from(str[, encoding])
 * Buffer.from(array)
 * Buffer.from(buffer)
 * Buffer.from(arrayBuffer[, byteOffset[, length]])
 **/
Buffer.from = function (value, encodingOrOffset, length) {
  return from(value, encodingOrOffset, length)
}

// Note: Change prototype *after* Buffer.from is defined to workaround Chrome bug:
// https://github.com/feross/buffer/pull/148
Buffer.prototype.__proto__ = Uint8Array.prototype
Buffer.__proto__ = Uint8Array

function assertSize (size) {
  if (typeof size !== 'number') {
    throw new TypeError('"size" argument must be of type number')
  } else if (size < 0) {
    throw new RangeError('The value "' + size + '" is invalid for option "size"')
  }
}

function alloc (size, fill, encoding) {
  assertSize(size)
  if (size <= 0) {
    return createBuffer(size)
  }
  if (fill !== undefined) {
    // Only pay attention to encoding if it's a string. This
    // prevents accidentally sending in a number that would
    // be interpretted as a start offset.
    return typeof encoding === 'string'
      ? createBuffer(size).fill(fill, encoding)
      : createBuffer(size).fill(fill)
  }
  return createBuffer(size)
}

/**
 * Creates a new filled Buffer instance.
 * alloc(size[, fill[, encoding]])
 **/
Buffer.alloc = function (size, fill, encoding) {
  return alloc(size, fill, encoding)
}

function allocUnsafe (size) {
  assertSize(size)
  return createBuffer(size < 0 ? 0 : checked(size) | 0)
}

/**
 * Equivalent to Buffer(num), by default creates a non-zero-filled Buffer instance.
 * */
Buffer.allocUnsafe = function (size) {
  return allocUnsafe(size)
}
/**
 * Equivalent to SlowBuffer(num), by default creates a non-zero-filled Buffer instance.
 */
Buffer.allocUnsafeSlow = function (size) {
  return allocUnsafe(size)
}

function fromString (string, encoding) {
  if (typeof encoding !== 'string' || encoding === '') {
    encoding = 'utf8'
  }

  if (!Buffer.isEncoding(encoding)) {
    throw new TypeError('Unknown encoding: ' + encoding)
  }

  var length = byteLength(string, encoding) | 0
  var buf = createBuffer(length)

  var actual = buf.write(string, encoding)

  if (actual !== length) {
    // Writing a hex string, for example, that contains invalid characters will
    // cause everything after the first invalid character to be ignored. (e.g.
    // 'abxxcd' will be treated as 'ab')
    buf = buf.slice(0, actual)
  }

  return buf
}

function fromArrayLike (array) {
  var length = array.length < 0 ? 0 : checked(array.length) | 0
  var buf = createBuffer(length)
  for (var i = 0; i < length; i += 1) {
    buf[i] = array[i] & 255
  }
  return buf
}

function fromArrayBuffer (array, byteOffset, length) {
  if (byteOffset < 0 || array.byteLength < byteOffset) {
    throw new RangeError('"offset" is outside of buffer bounds')
  }

  if (array.byteLength < byteOffset + (length || 0)) {
    throw new RangeError('"length" is outside of buffer bounds')
  }

  var buf
  if (byteOffset === undefined && length === undefined) {
    buf = new Uint8Array(array)
  } else if (length === undefined) {
    buf = new Uint8Array(array, byteOffset)
  } else {
    buf = new Uint8Array(array, byteOffset, length)
  }

  // Return an augmented `Uint8Array` instance
  buf.__proto__ = Buffer.prototype
  return buf
}

function fromObject (obj) {
  if (Buffer.isBuffer(obj)) {
    var len = checked(obj.length) | 0
    var buf = createBuffer(len)

    if (buf.length === 0) {
      return buf
    }

    obj.copy(buf, 0, 0, len)
    return buf
  }

  if (obj.length !== undefined) {
    if (typeof obj.length !== 'number' || numberIsNaN(obj.length)) {
      return createBuffer(0)
    }
    return fromArrayLike(obj)
  }

  if (obj.type === 'Buffer' && Array.isArray(obj.data)) {
    return fromArrayLike(obj.data)
  }
}

function checked (length) {
  // Note: cannot use `length < K_MAX_LENGTH` here because that fails when
  // length is NaN (which is otherwise coerced to zero.)
  if (length >= K_MAX_LENGTH) {
    throw new RangeError('Attempt to allocate Buffer larger than maximum ' +
                         'size: 0x' + K_MAX_LENGTH.toString(16) + ' bytes')
  }
  return length | 0
}

function SlowBuffer (length) {
  if (+length != length) { // eslint-disable-line eqeqeq
    length = 0
  }
  return Buffer.alloc(+length)
}

Buffer.isBuffer = function isBuffer (b) {
  return b != null && b._isBuffer === true &&
    b !== Buffer.prototype // so Buffer.isBuffer(Buffer.prototype) will be false
}

Buffer.compare = function compare (a, b) {
  if (isInstance(a, Uint8Array)) a = Buffer.from(a, a.offset, a.byteLength)
  if (isInstance(b, Uint8Array)) b = Buffer.from(b, b.offset, b.byteLength)
  if (!Buffer.isBuffer(a) || !Buffer.isBuffer(b)) {
    throw new TypeError(
      'The "buf1", "buf2" arguments must be one of type Buffer or Uint8Array'
    )
  }

  if (a === b) return 0

  var x = a.length
  var y = b.length

  for (var i = 0, len = Math.min(x, y); i < len; ++i) {
    if (a[i] !== b[i]) {
      x = a[i]
      y = b[i]
      break
    }
  }

  if (x < y) return -1
  if (y < x) return 1
  return 0
}

Buffer.isEncoding = function isEncoding (encoding) {
  switch (String(encoding).toLowerCase()) {
    case 'hex':
    case 'utf8':
    case 'utf-8':
    case 'ascii':
    case 'latin1':
    case 'binary':
    case 'base64':
    case 'ucs2':
    case 'ucs-2':
    case 'utf16le':
    case 'utf-16le':
      return true
    default:
      return false
  }
}

Buffer.concat = function concat (list, length) {
  if (!Array.isArray(list)) {
    throw new TypeError('"list" argument must be an Array of Buffers')
  }

  if (list.length === 0) {
    return Buffer.alloc(0)
  }

  var i
  if (length === undefined) {
    length = 0
    for (i = 0; i < list.length; ++i) {
      length += list[i].length
    }
  }

  var buffer = Buffer.allocUnsafe(length)
  var pos = 0
  for (i = 0; i < list.length; ++i) {
    var buf = list[i]
    if (isInstance(buf, Uint8Array)) {
      buf = Buffer.from(buf)
    }
    if (!Buffer.isBuffer(buf)) {
      throw new TypeError('"list" argument must be an Array of Buffers')
    }
    buf.copy(buffer, pos)
    pos += buf.length
  }
  return buffer
}

function byteLength (string, encoding) {
  if (Buffer.isBuffer(string)) {
    return string.length
  }
  if (ArrayBuffer.isView(string) || isInstance(string, ArrayBuffer)) {
    return string.byteLength
  }
  if (typeof string !== 'string') {
    throw new TypeError(
      'The "string" argument must be one of type string, Buffer, or ArrayBuffer. ' +
      'Received type ' + typeof string
    )
  }

  var len = string.length
  var mustMatch = (arguments.length > 2 && arguments[2] === true)
  if (!mustMatch && len === 0) return 0

  // Use a for loop to avoid recursion
  var loweredCase = false
  for (;;) {
    switch (encoding) {
      case 'ascii':
      case 'latin1':
      case 'binary':
        return len
      case 'utf8':
      case 'utf-8':
        return utf8ToBytes(string).length
      case 'ucs2':
      case 'ucs-2':
      case 'utf16le':
      case 'utf-16le':
        return len * 2
      case 'hex':
        return len >>> 1
      case 'base64':
        return base64ToBytes(string).length
      default:
        if (loweredCase) {
          return mustMatch ? -1 : utf8ToBytes(string).length // assume utf8
        }
        encoding = ('' + encoding).toLowerCase()
        loweredCase = true
    }
  }
}
Buffer.byteLength = byteLength

function slowToString (encoding, start, end) {
  var loweredCase = false

  // No need to verify that "this.length <= MAX_UINT32" since it's a read-only
  // property of a typed array.

  // This behaves neither like String nor Uint8Array in that we set start/end
  // to their upper/lower bounds if the value passed is out of range.
  // undefined is handled specially as per ECMA-262 6th Edition,
  // Section 13.3.3.7 Runtime Semantics: KeyedBindingInitialization.
  if (start === undefined || start < 0) {
    start = 0
  }
  // Return early if start > this.length. Done here to prevent potential uint32
  // coercion fail below.
  if (start > this.length) {
    return ''
  }

  if (end === undefined || end > this.length) {
    end = this.length
  }

  if (end <= 0) {
    return ''
  }

  // Force coersion to uint32. This will also coerce falsey/NaN values to 0.
  end >>>= 0
  start >>>= 0

  if (end <= start) {
    return ''
  }

  if (!encoding) encoding = 'utf8'

  while (true) {
    switch (encoding) {
      case 'hex':
        return hexSlice(this, start, end)

      case 'utf8':
      case 'utf-8':
        return utf8Slice(this, start, end)

      case 'ascii':
        return asciiSlice(this, start, end)

      case 'latin1':
      case 'binary':
        return latin1Slice(this, start, end)

      case 'base64':
        return base64Slice(this, start, end)

      case 'ucs2':
      case 'ucs-2':
      case 'utf16le':
      case 'utf-16le':
        return utf16leSlice(this, start, end)

      default:
        if (loweredCase) throw new TypeError('Unknown encoding: ' + encoding)
        encoding = (encoding + '').toLowerCase()
        loweredCase = true
    }
  }
}

// This property is used by `Buffer.isBuffer` (and the `is-buffer` npm package)
// to detect a Buffer instance. It's not possible to use `instanceof Buffer`
// reliably in a browserify context because there could be multiple different
// copies of the 'buffer' package in use. This method works even for Buffer
// instances that were created from another copy of the `buffer` package.
// See: https://github.com/feross/buffer/issues/154
Buffer.prototype._isBuffer = true

function swap (b, n, m) {
  var i = b[n]
  b[n] = b[m]
  b[m] = i
}

Buffer.prototype.swap16 = function swap16 () {
  var len = this.length
  if (len % 2 !== 0) {
    throw new RangeError('Buffer size must be a multiple of 16-bits')
  }
  for (var i = 0; i < len; i += 2) {
    swap(this, i, i + 1)
  }
  return this
}

Buffer.prototype.swap32 = function swap32 () {
  var len = this.length
  if (len % 4 !== 0) {
    throw new RangeError('Buffer size must be a multiple of 32-bits')
  }
  for (var i = 0; i < len; i += 4) {
    swap(this, i, i + 3)
    swap(this, i + 1, i + 2)
  }
  return this
}

Buffer.prototype.swap64 = function swap64 () {
  var len = this.length
  if (len % 8 !== 0) {
    throw new RangeError('Buffer size must be a multiple of 64-bits')
  }
  for (var i = 0; i < len; i += 8) {
    swap(this, i, i + 7)
    swap(this, i + 1, i + 6)
    swap(this, i + 2, i + 5)
    swap(this, i + 3, i + 4)
  }
  return this
}

Buffer.prototype.toString = function toString () {
  var length = this.length
  if (length === 0) return ''
  if (arguments.length === 0) return utf8Slice(this, 0, length)
  return slowToString.apply(this, arguments)
}

Buffer.prototype.toLocaleString = Buffer.prototype.toString

Buffer.prototype.equals = function equals (b) {
  if (!Buffer.isBuffer(b)) throw new TypeError('Argument must be a Buffer')
  if (this === b) return true
  return Buffer.compare(this, b) === 0
}

Buffer.prototype.inspect = function inspect () {
  var str = ''
  var max = exports.INSPECT_MAX_BYTES
  str = this.toString('hex', 0, max).replace(/(.{2})/g, '$1 ').trim()
  if (this.length > max) str += ' ... '
  return '<Buffer ' + str + '>'
}

Buffer.prototype.compare = function compare (target, start, end, thisStart, thisEnd) {
  if (isInstance(target, Uint8Array)) {
    target = Buffer.from(target, target.offset, target.byteLength)
  }
  if (!Buffer.isBuffer(target)) {
    throw new TypeError(
      'The "target" argument must be one of type Buffer or Uint8Array. ' +
      'Received type ' + (typeof target)
    )
  }

  if (start === undefined) {
    start = 0
  }
  if (end === undefined) {
    end = target ? target.length : 0
  }
  if (thisStart === undefined) {
    thisStart = 0
  }
  if (thisEnd === undefined) {
    thisEnd = this.length
  }

  if (start < 0 || end > target.length || thisStart < 0 || thisEnd > this.length) {
    throw new RangeError('out of range index')
  }

  if (thisStart >= thisEnd && start >= end) {
    return 0
  }
  if (thisStart >= thisEnd) {
    return -1
  }
  if (start >= end) {
    return 1
  }

  start >>>= 0
  end >>>= 0
  thisStart >>>= 0
  thisEnd >>>= 0

  if (this === target) return 0

  var x = thisEnd - thisStart
  var y = end - start
  var len = Math.min(x, y)

  var thisCopy = this.slice(thisStart, thisEnd)
  var targetCopy = target.slice(start, end)

  for (var i = 0; i < len; ++i) {
    if (thisCopy[i] !== targetCopy[i]) {
      x = thisCopy[i]
      y = targetCopy[i]
      break
    }
  }

  if (x < y) return -1
  if (y < x) return 1
  return 0
}

// Finds either the first index of `val` in `buffer` at offset >= `byteOffset`,
// OR the last index of `val` in `buffer` at offset <= `byteOffset`.
//
// Arguments:
// - buffer - a Buffer to search
// - val - a string, Buffer, or number
// - byteOffset - an index into `buffer`; will be clamped to an int32
// - encoding - an optional encoding, relevant is val is a string
// - dir - true for indexOf, false for lastIndexOf
function bidirectionalIndexOf (buffer, val, byteOffset, encoding, dir) {
  // Empty buffer means no match
  if (buffer.length === 0) return -1

  // Normalize byteOffset
  if (typeof byteOffset === 'string') {
    encoding = byteOffset
    byteOffset = 0
  } else if (byteOffset > 0x7fffffff) {
    byteOffset = 0x7fffffff
  } else if (byteOffset < -0x80000000) {
    byteOffset = -0x80000000
  }
  byteOffset = +byteOffset // Coerce to Number.
  if (numberIsNaN(byteOffset)) {
    // byteOffset: it it's undefined, null, NaN, "foo", etc, search whole buffer
    byteOffset = dir ? 0 : (buffer.length - 1)
  }

  // Normalize byteOffset: negative offsets start from the end of the buffer
  if (byteOffset < 0) byteOffset = buffer.length + byteOffset
  if (byteOffset >= buffer.length) {
    if (dir) return -1
    else byteOffset = buffer.length - 1
  } else if (byteOffset < 0) {
    if (dir) byteOffset = 0
    else return -1
  }

  // Normalize val
  if (typeof val === 'string') {
    val = Buffer.from(val, encoding)
  }

  // Finally, search either indexOf (if dir is true) or lastIndexOf
  if (Buffer.isBuffer(val)) {
    // Special case: looking for empty string/buffer always fails
    if (val.length === 0) {
      return -1
    }
    return arrayIndexOf(buffer, val, byteOffset, encoding, dir)
  } else if (typeof val === 'number') {
    val = val & 0xFF // Search for a byte value [0-255]
    if (typeof Uint8Array.prototype.indexOf === 'function') {
      if (dir) {
        return Uint8Array.prototype.indexOf.call(buffer, val, byteOffset)
      } else {
        return Uint8Array.prototype.lastIndexOf.call(buffer, val, byteOffset)
      }
    }
    return arrayIndexOf(buffer, [ val ], byteOffset, encoding, dir)
  }

  throw new TypeError('val must be string, number or Buffer')
}

function arrayIndexOf (arr, val, byteOffset, encoding, dir) {
  var indexSize = 1
  var arrLength = arr.length
  var valLength = val.length

  if (encoding !== undefined) {
    encoding = String(encoding).toLowerCase()
    if (encoding === 'ucs2' || encoding === 'ucs-2' ||
        encoding === 'utf16le' || encoding === 'utf-16le') {
      if (arr.length < 2 || val.length < 2) {
        return -1
      }
      indexSize = 2
      arrLength /= 2
      valLength /= 2
      byteOffset /= 2
    }
  }

  function read (buf, i) {
    if (indexSize === 1) {
      return buf[i]
    } else {
      return buf.readUInt16BE(i * indexSize)
    }
  }

  var i
  if (dir) {
    var foundIndex = -1
    for (i = byteOffset; i < arrLength; i++) {
      if (read(arr, i) === read(val, foundIndex === -1 ? 0 : i - foundIndex)) {
        if (foundIndex === -1) foundIndex = i
        if (i - foundIndex + 1 === valLength) return foundIndex * indexSize
      } else {
        if (foundIndex !== -1) i -= i - foundIndex
        foundIndex = -1
      }
    }
  } else {
    if (byteOffset + valLength > arrLength) byteOffset = arrLength - valLength
    for (i = byteOffset; i >= 0; i--) {
      var found = true
      for (var j = 0; j < valLength; j++) {
        if (read(arr, i + j) !== read(val, j)) {
          found = false
          break
        }
      }
      if (found) return i
    }
  }

  return -1
}

Buffer.prototype.includes = function includes (val, byteOffset, encoding) {
  return this.indexOf(val, byteOffset, encoding) !== -1
}

Buffer.prototype.indexOf = function indexOf (val, byteOffset, encoding) {
  return bidirectionalIndexOf(this, val, byteOffset, encoding, true)
}

Buffer.prototype.lastIndexOf = function lastIndexOf (val, byteOffset, encoding) {
  return bidirectionalIndexOf(this, val, byteOffset, encoding, false)
}

function hexWrite (buf, string, offset, length) {
  offset = Number(offset) || 0
  var remaining = buf.length - offset
  if (!length) {
    length = remaining
  } else {
    length = Number(length)
    if (length > remaining) {
      length = remaining
    }
  }

  var strLen = string.length

  if (length > strLen / 2) {
    length = strLen / 2
  }
  for (var i = 0; i < length; ++i) {
    var parsed = parseInt(string.substr(i * 2, 2), 16)
    if (numberIsNaN(parsed)) return i
    buf[offset + i] = parsed
  }
  return i
}

function utf8Write (buf, string, offset, length) {
  return blitBuffer(utf8ToBytes(string, buf.length - offset), buf, offset, length)
}

function asciiWrite (buf, string, offset, length) {
  return blitBuffer(asciiToBytes(string), buf, offset, length)
}

function latin1Write (buf, string, offset, length) {
  return asciiWrite(buf, string, offset, length)
}

function base64Write (buf, string, offset, length) {
  return blitBuffer(base64ToBytes(string), buf, offset, length)
}

function ucs2Write (buf, string, offset, length) {
  return blitBuffer(utf16leToBytes(string, buf.length - offset), buf, offset, length)
}

Buffer.prototype.write = function write (string, offset, length, encoding) {
  // Buffer#write(string)
  if (offset === undefined) {
    encoding = 'utf8'
    length = this.length
    offset = 0
  // Buffer#write(string, encoding)
  } else if (length === undefined && typeof offset === 'string') {
    encoding = offset
    length = this.length
    offset = 0
  // Buffer#write(string, offset[, length][, encoding])
  } else if (isFinite(offset)) {
    offset = offset >>> 0
    if (isFinite(length)) {
      length = length >>> 0
      if (encoding === undefined) encoding = 'utf8'
    } else {
      encoding = length
      length = undefined
    }
  } else {
    throw new Error(
      'Buffer.write(string, encoding, offset[, length]) is no longer supported'
    )
  }

  var remaining = this.length - offset
  if (length === undefined || length > remaining) length = remaining

  if ((string.length > 0 && (length < 0 || offset < 0)) || offset > this.length) {
    throw new RangeError('Attempt to write outside buffer bounds')
  }

  if (!encoding) encoding = 'utf8'

  var loweredCase = false
  for (;;) {
    switch (encoding) {
      case 'hex':
        return hexWrite(this, string, offset, length)

      case 'utf8':
      case 'utf-8':
        return utf8Write(this, string, offset, length)

      case 'ascii':
        return asciiWrite(this, string, offset, length)

      case 'latin1':
      case 'binary':
        return latin1Write(this, string, offset, length)

      case 'base64':
        // Warning: maxLength not taken into account in base64Write
        return base64Write(this, string, offset, length)

      case 'ucs2':
      case 'ucs-2':
      case 'utf16le':
      case 'utf-16le':
        return ucs2Write(this, string, offset, length)

      default:
        if (loweredCase) throw new TypeError('Unknown encoding: ' + encoding)
        encoding = ('' + encoding).toLowerCase()
        loweredCase = true
    }
  }
}

Buffer.prototype.toJSON = function toJSON () {
  return {
    type: 'Buffer',
    data: Array.prototype.slice.call(this._arr || this, 0)
  }
}

function base64Slice (buf, start, end) {
  if (start === 0 && end === buf.length) {
    return base64.fromByteArray(buf)
  } else {
    return base64.fromByteArray(buf.slice(start, end))
  }
}

function utf8Slice (buf, start, end) {
  end = Math.min(buf.length, end)
  var res = []

  var i = start
  while (i < end) {
    var firstByte = buf[i]
    var codePoint = null
    var bytesPerSequence = (firstByte > 0xEF) ? 4
      : (firstByte > 0xDF) ? 3
        : (firstByte > 0xBF) ? 2
          : 1

    if (i + bytesPerSequence <= end) {
      var secondByte, thirdByte, fourthByte, tempCodePoint

      switch (bytesPerSequence) {
        case 1:
          if (firstByte < 0x80) {
            codePoint = firstByte
          }
          break
        case 2:
          secondByte = buf[i + 1]
          if ((secondByte & 0xC0) === 0x80) {
            tempCodePoint = (firstByte & 0x1F) << 0x6 | (secondByte & 0x3F)
            if (tempCodePoint > 0x7F) {
              codePoint = tempCodePoint
            }
          }
          break
        case 3:
          secondByte = buf[i + 1]
          thirdByte = buf[i + 2]
          if ((secondByte & 0xC0) === 0x80 && (thirdByte & 0xC0) === 0x80) {
            tempCodePoint = (firstByte & 0xF) << 0xC | (secondByte & 0x3F) << 0x6 | (thirdByte & 0x3F)
            if (tempCodePoint > 0x7FF && (tempCodePoint < 0xD800 || tempCodePoint > 0xDFFF)) {
              codePoint = tempCodePoint
            }
          }
          break
        case 4:
          secondByte = buf[i + 1]
          thirdByte = buf[i + 2]
          fourthByte = buf[i + 3]
          if ((secondByte & 0xC0) === 0x80 && (thirdByte & 0xC0) === 0x80 && (fourthByte & 0xC0) === 0x80) {
            tempCodePoint = (firstByte & 0xF) << 0x12 | (secondByte & 0x3F) << 0xC | (thirdByte & 0x3F) << 0x6 | (fourthByte & 0x3F)
            if (tempCodePoint > 0xFFFF && tempCodePoint < 0x110000) {
              codePoint = tempCodePoint
            }
          }
      }
    }

    if (codePoint === null) {
      // we did not generate a valid codePoint so insert a
      // replacement char (U+FFFD) and advance only 1 byte
      codePoint = 0xFFFD
      bytesPerSequence = 1
    } else if (codePoint > 0xFFFF) {
      // encode to utf16 (surrogate pair dance)
      codePoint -= 0x10000
      res.push(codePoint >>> 10 & 0x3FF | 0xD800)
      codePoint = 0xDC00 | codePoint & 0x3FF
    }

    res.push(codePoint)
    i += bytesPerSequence
  }

  return decodeCodePointsArray(res)
}

// Based on http://stackoverflow.com/a/22747272/680742, the browser with
// the lowest limit is Chrome, with 0x10000 args.
// We go 1 magnitude less, for safety
var MAX_ARGUMENTS_LENGTH = 0x1000

function decodeCodePointsArray (codePoints) {
  var len = codePoints.length
  if (len <= MAX_ARGUMENTS_LENGTH) {
    return String.fromCharCode.apply(String, codePoints) // avoid extra slice()
  }

  // Decode in chunks to avoid "call stack size exceeded".
  var res = ''
  var i = 0
  while (i < len) {
    res += String.fromCharCode.apply(
      String,
      codePoints.slice(i, i += MAX_ARGUMENTS_LENGTH)
    )
  }
  return res
}

function asciiSlice (buf, start, end) {
  var ret = ''
  end = Math.min(buf.length, end)

  for (var i = start; i < end; ++i) {
    ret += String.fromCharCode(buf[i] & 0x7F)
  }
  return ret
}

function latin1Slice (buf, start, end) {
  var ret = ''
  end = Math.min(buf.length, end)

  for (var i = start; i < end; ++i) {
    ret += String.fromCharCode(buf[i])
  }
  return ret
}

function hexSlice (buf, start, end) {
  var len = buf.length

  if (!start || start < 0) start = 0
  if (!end || end < 0 || end > len) end = len

  var out = ''
  for (var i = start; i < end; ++i) {
    out += toHex(buf[i])
  }
  return out
}

function utf16leSlice (buf, start, end) {
  var bytes = buf.slice(start, end)
  var res = ''
  for (var i = 0; i < bytes.length; i += 2) {
    res += String.fromCharCode(bytes[i] + (bytes[i + 1] * 256))
  }
  return res
}

Buffer.prototype.slice = function slice (start, end) {
  var len = this.length
  start = ~~start
  end = end === undefined ? len : ~~end

  if (start < 0) {
    start += len
    if (start < 0) start = 0
  } else if (start > len) {
    start = len
  }

  if (end < 0) {
    end += len
    if (end < 0) end = 0
  } else if (end > len) {
    end = len
  }

  if (end < start) end = start

  var newBuf = this.subarray(start, end)
  // Return an augmented `Uint8Array` instance
  newBuf.__proto__ = Buffer.prototype
  return newBuf
}

/*
 * Need to make sure that buffer isn't trying to write out of bounds.
 */
function checkOffset (offset, ext, length) {
  if ((offset % 1) !== 0 || offset < 0) throw new RangeError('offset is not uint')
  if (offset + ext > length) throw new RangeError('Trying to access beyond buffer length')
}

Buffer.prototype.readUIntLE = function readUIntLE (offset, byteLength, noAssert) {
  offset = offset >>> 0
  byteLength = byteLength >>> 0
  if (!noAssert) checkOffset(offset, byteLength, this.length)

  var val = this[offset]
  var mul = 1
  var i = 0
  while (++i < byteLength && (mul *= 0x100)) {
    val += this[offset + i] * mul
  }

  return val
}

Buffer.prototype.readUIntBE = function readUIntBE (offset, byteLength, noAssert) {
  offset = offset >>> 0
  byteLength = byteLength >>> 0
  if (!noAssert) {
    checkOffset(offset, byteLength, this.length)
  }

  var val = this[offset + --byteLength]
  var mul = 1
  while (byteLength > 0 && (mul *= 0x100)) {
    val += this[offset + --byteLength] * mul
  }

  return val
}

Buffer.prototype.readUInt8 = function readUInt8 (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 1, this.length)
  return this[offset]
}

Buffer.prototype.readUInt16LE = function readUInt16LE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 2, this.length)
  return this[offset] | (this[offset + 1] << 8)
}

Buffer.prototype.readUInt16BE = function readUInt16BE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 2, this.length)
  return (this[offset] << 8) | this[offset + 1]
}

Buffer.prototype.readUInt32LE = function readUInt32LE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 4, this.length)

  return ((this[offset]) |
      (this[offset + 1] << 8) |
      (this[offset + 2] << 16)) +
      (this[offset + 3] * 0x1000000)
}

Buffer.prototype.readUInt32BE = function readUInt32BE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 4, this.length)

  return (this[offset] * 0x1000000) +
    ((this[offset + 1] << 16) |
    (this[offset + 2] << 8) |
    this[offset + 3])
}

Buffer.prototype.readIntLE = function readIntLE (offset, byteLength, noAssert) {
  offset = offset >>> 0
  byteLength = byteLength >>> 0
  if (!noAssert) checkOffset(offset, byteLength, this.length)

  var val = this[offset]
  var mul = 1
  var i = 0
  while (++i < byteLength && (mul *= 0x100)) {
    val += this[offset + i] * mul
  }
  mul *= 0x80

  if (val >= mul) val -= Math.pow(2, 8 * byteLength)

  return val
}

Buffer.prototype.readIntBE = function readIntBE (offset, byteLength, noAssert) {
  offset = offset >>> 0
  byteLength = byteLength >>> 0
  if (!noAssert) checkOffset(offset, byteLength, this.length)

  var i = byteLength
  var mul = 1
  var val = this[offset + --i]
  while (i > 0 && (mul *= 0x100)) {
    val += this[offset + --i] * mul
  }
  mul *= 0x80

  if (val >= mul) val -= Math.pow(2, 8 * byteLength)

  return val
}

Buffer.prototype.readInt8 = function readInt8 (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 1, this.length)
  if (!(this[offset] & 0x80)) return (this[offset])
  return ((0xff - this[offset] + 1) * -1)
}

Buffer.prototype.readInt16LE = function readInt16LE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 2, this.length)
  var val = this[offset] | (this[offset + 1] << 8)
  return (val & 0x8000) ? val | 0xFFFF0000 : val
}

Buffer.prototype.readInt16BE = function readInt16BE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 2, this.length)
  var val = this[offset + 1] | (this[offset] << 8)
  return (val & 0x8000) ? val | 0xFFFF0000 : val
}

Buffer.prototype.readInt32LE = function readInt32LE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 4, this.length)

  return (this[offset]) |
    (this[offset + 1] << 8) |
    (this[offset + 2] << 16) |
    (this[offset + 3] << 24)
}

Buffer.prototype.readInt32BE = function readInt32BE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 4, this.length)

  return (this[offset] << 24) |
    (this[offset + 1] << 16) |
    (this[offset + 2] << 8) |
    (this[offset + 3])
}

Buffer.prototype.readFloatLE = function readFloatLE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 4, this.length)
  return ieee754.read(this, offset, true, 23, 4)
}

Buffer.prototype.readFloatBE = function readFloatBE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 4, this.length)
  return ieee754.read(this, offset, false, 23, 4)
}

Buffer.prototype.readDoubleLE = function readDoubleLE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 8, this.length)
  return ieee754.read(this, offset, true, 52, 8)
}

Buffer.prototype.readDoubleBE = function readDoubleBE (offset, noAssert) {
  offset = offset >>> 0
  if (!noAssert) checkOffset(offset, 8, this.length)
  return ieee754.read(this, offset, false, 52, 8)
}

function checkInt (buf, value, offset, ext, max, min) {
  if (!Buffer.isBuffer(buf)) throw new TypeError('"buffer" argument must be a Buffer instance')
  if (value > max || value < min) throw new RangeError('"value" argument is out of bounds')
  if (offset + ext > buf.length) throw new RangeError('Index out of range')
}

Buffer.prototype.writeUIntLE = function writeUIntLE (value, offset, byteLength, noAssert) {
  value = +value
  offset = offset >>> 0
  byteLength = byteLength >>> 0
  if (!noAssert) {
    var maxBytes = Math.pow(2, 8 * byteLength) - 1
    checkInt(this, value, offset, byteLength, maxBytes, 0)
  }

  var mul = 1
  var i = 0
  this[offset] = value & 0xFF
  while (++i < byteLength && (mul *= 0x100)) {
    this[offset + i] = (value / mul) & 0xFF
  }

  return offset + byteLength
}

Buffer.prototype.writeUIntBE = function writeUIntBE (value, offset, byteLength, noAssert) {
  value = +value
  offset = offset >>> 0
  byteLength = byteLength >>> 0
  if (!noAssert) {
    var maxBytes = Math.pow(2, 8 * byteLength) - 1
    checkInt(this, value, offset, byteLength, maxBytes, 0)
  }

  var i = byteLength - 1
  var mul = 1
  this[offset + i] = value & 0xFF
  while (--i >= 0 && (mul *= 0x100)) {
    this[offset + i] = (value / mul) & 0xFF
  }

  return offset + byteLength
}

Buffer.prototype.writeUInt8 = function writeUInt8 (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 1, 0xff, 0)
  this[offset] = (value & 0xff)
  return offset + 1
}

Buffer.prototype.writeUInt16LE = function writeUInt16LE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 2, 0xffff, 0)
  this[offset] = (value & 0xff)
  this[offset + 1] = (value >>> 8)
  return offset + 2
}

Buffer.prototype.writeUInt16BE = function writeUInt16BE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 2, 0xffff, 0)
  this[offset] = (value >>> 8)
  this[offset + 1] = (value & 0xff)
  return offset + 2
}

Buffer.prototype.writeUInt32LE = function writeUInt32LE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 4, 0xffffffff, 0)
  this[offset + 3] = (value >>> 24)
  this[offset + 2] = (value >>> 16)
  this[offset + 1] = (value >>> 8)
  this[offset] = (value & 0xff)
  return offset + 4
}

Buffer.prototype.writeUInt32BE = function writeUInt32BE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 4, 0xffffffff, 0)
  this[offset] = (value >>> 24)
  this[offset + 1] = (value >>> 16)
  this[offset + 2] = (value >>> 8)
  this[offset + 3] = (value & 0xff)
  return offset + 4
}

Buffer.prototype.writeIntLE = function writeIntLE (value, offset, byteLength, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) {
    var limit = Math.pow(2, (8 * byteLength) - 1)

    checkInt(this, value, offset, byteLength, limit - 1, -limit)
  }

  var i = 0
  var mul = 1
  var sub = 0
  this[offset] = value & 0xFF
  while (++i < byteLength && (mul *= 0x100)) {
    if (value < 0 && sub === 0 && this[offset + i - 1] !== 0) {
      sub = 1
    }
    this[offset + i] = ((value / mul) >> 0) - sub & 0xFF
  }

  return offset + byteLength
}

Buffer.prototype.writeIntBE = function writeIntBE (value, offset, byteLength, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) {
    var limit = Math.pow(2, (8 * byteLength) - 1)

    checkInt(this, value, offset, byteLength, limit - 1, -limit)
  }

  var i = byteLength - 1
  var mul = 1
  var sub = 0
  this[offset + i] = value & 0xFF
  while (--i >= 0 && (mul *= 0x100)) {
    if (value < 0 && sub === 0 && this[offset + i + 1] !== 0) {
      sub = 1
    }
    this[offset + i] = ((value / mul) >> 0) - sub & 0xFF
  }

  return offset + byteLength
}

Buffer.prototype.writeInt8 = function writeInt8 (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 1, 0x7f, -0x80)
  if (value < 0) value = 0xff + value + 1
  this[offset] = (value & 0xff)
  return offset + 1
}

Buffer.prototype.writeInt16LE = function writeInt16LE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 2, 0x7fff, -0x8000)
  this[offset] = (value & 0xff)
  this[offset + 1] = (value >>> 8)
  return offset + 2
}

Buffer.prototype.writeInt16BE = function writeInt16BE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 2, 0x7fff, -0x8000)
  this[offset] = (value >>> 8)
  this[offset + 1] = (value & 0xff)
  return offset + 2
}

Buffer.prototype.writeInt32LE = function writeInt32LE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 4, 0x7fffffff, -0x80000000)
  this[offset] = (value & 0xff)
  this[offset + 1] = (value >>> 8)
  this[offset + 2] = (value >>> 16)
  this[offset + 3] = (value >>> 24)
  return offset + 4
}

Buffer.prototype.writeInt32BE = function writeInt32BE (value, offset, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) checkInt(this, value, offset, 4, 0x7fffffff, -0x80000000)
  if (value < 0) value = 0xffffffff + value + 1
  this[offset] = (value >>> 24)
  this[offset + 1] = (value >>> 16)
  this[offset + 2] = (value >>> 8)
  this[offset + 3] = (value & 0xff)
  return offset + 4
}

function checkIEEE754 (buf, value, offset, ext, max, min) {
  if (offset + ext > buf.length) throw new RangeError('Index out of range')
  if (offset < 0) throw new RangeError('Index out of range')
}

function writeFloat (buf, value, offset, littleEndian, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) {
    checkIEEE754(buf, value, offset, 4, 3.4028234663852886e+38, -3.4028234663852886e+38)
  }
  ieee754.write(buf, value, offset, littleEndian, 23, 4)
  return offset + 4
}

Buffer.prototype.writeFloatLE = function writeFloatLE (value, offset, noAssert) {
  return writeFloat(this, value, offset, true, noAssert)
}

Buffer.prototype.writeFloatBE = function writeFloatBE (value, offset, noAssert) {
  return writeFloat(this, value, offset, false, noAssert)
}

function writeDouble (buf, value, offset, littleEndian, noAssert) {
  value = +value
  offset = offset >>> 0
  if (!noAssert) {
    checkIEEE754(buf, value, offset, 8, 1.7976931348623157E+308, -1.7976931348623157E+308)
  }
  ieee754.write(buf, value, offset, littleEndian, 52, 8)
  return offset + 8
}

Buffer.prototype.writeDoubleLE = function writeDoubleLE (value, offset, noAssert) {
  return writeDouble(this, value, offset, true, noAssert)
}

Buffer.prototype.writeDoubleBE = function writeDoubleBE (value, offset, noAssert) {
  return writeDouble(this, value, offset, false, noAssert)
}

// copy(targetBuffer, targetStart=0, sourceStart=0, sourceEnd=buffer.length)
Buffer.prototype.copy = function copy (target, targetStart, start, end) {
  if (!Buffer.isBuffer(target)) throw new TypeError('argument should be a Buffer')
  if (!start) start = 0
  if (!end && end !== 0) end = this.length
  if (targetStart >= target.length) targetStart = target.length
  if (!targetStart) targetStart = 0
  if (end > 0 && end < start) end = start

  // Copy 0 bytes; we're done
  if (end === start) return 0
  if (target.length === 0 || this.length === 0) return 0

  // Fatal error conditions
  if (targetStart < 0) {
    throw new RangeError('targetStart out of bounds')
  }
  if (start < 0 || start >= this.length) throw new RangeError('Index out of range')
  if (end < 0) throw new RangeError('sourceEnd out of bounds')

  // Are we oob?
  if (end > this.length) end = this.length
  if (target.length - targetStart < end - start) {
    end = target.length - targetStart + start
  }

  var len = end - start

  if (this === target && typeof Uint8Array.prototype.copyWithin === 'function') {
    // Use built-in when available, missing from IE11
    this.copyWithin(targetStart, start, end)
  } else if (this === target && start < targetStart && targetStart < end) {
    // descending copy from end
    for (var i = len - 1; i >= 0; --i) {
      target[i + targetStart] = this[i + start]
    }
  } else {
    Uint8Array.prototype.set.call(
      target,
      this.subarray(start, end),
      targetStart
    )
  }

  return len
}

// Usage:
//    buffer.fill(number[, offset[, end]])
//    buffer.fill(buffer[, offset[, end]])
//    buffer.fill(string[, offset[, end]][, encoding])
Buffer.prototype.fill = function fill (val, start, end, encoding) {
  // Handle string cases:
  if (typeof val === 'string') {
    if (typeof start === 'string') {
      encoding = start
      start = 0
      end = this.length
    } else if (typeof end === 'string') {
      encoding = end
      end = this.length
    }
    if (encoding !== undefined && typeof encoding !== 'string') {
      throw new TypeError('encoding must be a string')
    }
    if (typeof encoding === 'string' && !Buffer.isEncoding(encoding)) {
      throw new TypeError('Unknown encoding: ' + encoding)
    }
    if (val.length === 1) {
      var code = val.charCodeAt(0)
      if ((encoding === 'utf8' && code < 128) ||
          encoding === 'latin1') {
        // Fast path: If `val` fits into a single byte, use that numeric value.
        val = code
      }
    }
  } else if (typeof val === 'number') {
    val = val & 255
  }

  // Invalid ranges are not set to a default, so can range check early.
  if (start < 0 || this.length < start || this.length < end) {
    throw new RangeError('Out of range index')
  }

  if (end <= start) {
    return this
  }

  start = start >>> 0
  end = end === undefined ? this.length : end >>> 0

  if (!val) val = 0

  var i
  if (typeof val === 'number') {
    for (i = start; i < end; ++i) {
      this[i] = val
    }
  } else {
    var bytes = Buffer.isBuffer(val)
      ? val
      : Buffer.from(val, encoding)
    var len = bytes.length
    if (len === 0) {
      throw new TypeError('The value "' + val +
        '" is invalid for argument "value"')
    }
    for (i = 0; i < end - start; ++i) {
      this[i + start] = bytes[i % len]
    }
  }

  return this
}

// HELPER FUNCTIONS
// ================

var INVALID_BASE64_RE = /[^+/0-9A-Za-z-_]/g

function base64clean (str) {
  // Node takes equal signs as end of the Base64 encoding
  str = str.split('=')[0]
  // Node strips out invalid characters like \n and \t from the string, base64-js does not
  str = str.trim().replace(INVALID_BASE64_RE, '')
  // Node converts strings with length < 2 to ''
  if (str.length < 2) return ''
  // Node allows for non-padded base64 strings (missing trailing ===), base64-js does not
  while (str.length % 4 !== 0) {
    str = str + '='
  }
  return str
}

function toHex (n) {
  if (n < 16) return '0' + n.toString(16)
  return n.toString(16)
}

function utf8ToBytes (string, units) {
  units = units || Infinity
  var codePoint
  var length = string.length
  var leadSurrogate = null
  var bytes = []

  for (var i = 0; i < length; ++i) {
    codePoint = string.charCodeAt(i)

    // is surrogate component
    if (codePoint > 0xD7FF && codePoint < 0xE000) {
      // last char was a lead
      if (!leadSurrogate) {
        // no lead yet
        if (codePoint > 0xDBFF) {
          // unexpected trail
          if ((units -= 3) > -1) bytes.push(0xEF, 0xBF, 0xBD)
          continue
        } else if (i + 1 === length) {
          // unpaired lead
          if ((units -= 3) > -1) bytes.push(0xEF, 0xBF, 0xBD)
          continue
        }

        // valid lead
        leadSurrogate = codePoint

        continue
      }

      // 2 leads in a row
      if (codePoint < 0xDC00) {
        if ((units -= 3) > -1) bytes.push(0xEF, 0xBF, 0xBD)
        leadSurrogate = codePoint
        continue
      }

      // valid surrogate pair
      codePoint = (leadSurrogate - 0xD800 << 10 | codePoint - 0xDC00) + 0x10000
    } else if (leadSurrogate) {
      // valid bmp char, but last char was a lead
      if ((units -= 3) > -1) bytes.push(0xEF, 0xBF, 0xBD)
    }

    leadSurrogate = null

    // encode utf8
    if (codePoint < 0x80) {
      if ((units -= 1) < 0) break
      bytes.push(codePoint)
    } else if (codePoint < 0x800) {
      if ((units -= 2) < 0) break
      bytes.push(
        codePoint >> 0x6 | 0xC0,
        codePoint & 0x3F | 0x80
      )
    } else if (codePoint < 0x10000) {
      if ((units -= 3) < 0) break
      bytes.push(
        codePoint >> 0xC | 0xE0,
        codePoint >> 0x6 & 0x3F | 0x80,
        codePoint & 0x3F | 0x80
      )
    } else if (codePoint < 0x110000) {
      if ((units -= 4) < 0) break
      bytes.push(
        codePoint >> 0x12 | 0xF0,
        codePoint >> 0xC & 0x3F | 0x80,
        codePoint >> 0x6 & 0x3F | 0x80,
        codePoint & 0x3F | 0x80
      )
    } else {
      throw new Error('Invalid code point')
    }
  }

  return bytes
}

function asciiToBytes (str) {
  var byteArray = []
  for (var i = 0; i < str.length; ++i) {
    // Node's code seems to be doing this and not & 0x7F..
    byteArray.push(str.charCodeAt(i) & 0xFF)
  }
  return byteArray
}

function utf16leToBytes (str, units) {
  var c, hi, lo
  var byteArray = []
  for (var i = 0; i < str.length; ++i) {
    if ((units -= 2) < 0) break

    c = str.charCodeAt(i)
    hi = c >> 8
    lo = c % 256
    byteArray.push(lo)
    byteArray.push(hi)
  }

  return byteArray
}

function base64ToBytes (str) {
  return base64.toByteArray(base64clean(str))
}

function blitBuffer (src, dst, offset, length) {
  for (var i = 0; i < length; ++i) {
    if ((i + offset >= dst.length) || (i >= src.length)) break
    dst[i + offset] = src[i]
  }
  return i
}

// ArrayBuffer or Uint8Array objects from other contexts (i.e. iframes) do not pass
// the `instanceof` check but they should be treated as of that type.
// See: https://github.com/feross/buffer/issues/166
function isInstance (obj, type) {
  return obj instanceof type ||
    (obj != null && obj.constructor != null && obj.constructor.name != null &&
      obj.constructor.name === type.name)
}
function numberIsNaN (obj) {
  // For IE11 support
  return obj !== obj // eslint-disable-line no-self-compare
}

}).call(this)}).call(this,require("buffer").Buffer)
},{"base64-js":2,"buffer":4,"ieee754":5}],5:[function(require,module,exports){
/*! ieee754. BSD-3-Clause License. Feross Aboukhadijeh <https://feross.org/opensource> */
exports.read = function (buffer, offset, isLE, mLen, nBytes) {
  var e, m
  var eLen = (nBytes * 8) - mLen - 1
  var eMax = (1 << eLen) - 1
  var eBias = eMax >> 1
  var nBits = -7
  var i = isLE ? (nBytes - 1) : 0
  var d = isLE ? -1 : 1
  var s = buffer[offset + i]

  i += d

  e = s & ((1 << (-nBits)) - 1)
  s >>= (-nBits)
  nBits += eLen
  for (; nBits > 0; e = (e * 256) + buffer[offset + i], i += d, nBits -= 8) {}

  m = e & ((1 << (-nBits)) - 1)
  e >>= (-nBits)
  nBits += mLen
  for (; nBits > 0; m = (m * 256) + buffer[offset + i], i += d, nBits -= 8) {}

  if (e === 0) {
    e = 1 - eBias
  } else if (e === eMax) {
    return m ? NaN : ((s ? -1 : 1) * Infinity)
  } else {
    m = m + Math.pow(2, mLen)
    e = e - eBias
  }
  return (s ? -1 : 1) * m * Math.pow(2, e - mLen)
}

exports.write = function (buffer, value, offset, isLE, mLen, nBytes) {
  var e, m, c
  var eLen = (nBytes * 8) - mLen - 1
  var eMax = (1 << eLen) - 1
  var eBias = eMax >> 1
  var rt = (mLen === 23 ? Math.pow(2, -24) - Math.pow(2, -77) : 0)
  var i = isLE ? 0 : (nBytes - 1)
  var d = isLE ? 1 : -1
  var s = value < 0 || (value === 0 && 1 / value < 0) ? 1 : 0

  value = Math.abs(value)

  if (isNaN(value) || value === Infinity) {
    m = isNaN(value) ? 1 : 0
    e = eMax
  } else {
    e = Math.floor(Math.log(value) / Math.LN2)
    if (value * (c = Math.pow(2, -e)) < 1) {
      e--
      c *= 2
    }
    if (e + eBias >= 1) {
      value += rt / c
    } else {
      value += rt * Math.pow(2, 1 - eBias)
    }
    if (value * c >= 2) {
      e++
      c /= 2
    }

    if (e + eBias >= eMax) {
      m = 0
      e = eMax
    } else if (e + eBias >= 1) {
      m = ((value * c) - 1) * Math.pow(2, mLen)
      e = e + eBias
    } else {
      m = value * Math.pow(2, eBias - 1) * Math.pow(2, mLen)
      e = 0
    }
  }

  for (; mLen >= 8; buffer[offset + i] = m & 0xff, i += d, m /= 256, mLen -= 8) {}

  e = (e << mLen) | m
  eLen += mLen
  for (; eLen > 0; buffer[offset + i] = e & 0xff, i += d, e /= 256, eLen -= 8) {}

  buffer[offset + i - d] |= s * 128
}

},{}],6:[function(require,module,exports){
// shim for using process in browser
var process = module.exports = {};

// cached from whatever global is present so that test runners that stub it
// don't break things.  But we need to wrap it in a try catch in case it is
// wrapped in strict mode code which doesn't define any globals.  It's inside a
// function because try/catches deoptimize in certain engines.

var cachedSetTimeout;
var cachedClearTimeout;

function defaultSetTimout() {
    throw new Error('setTimeout has not been defined');
}
function defaultClearTimeout () {
    throw new Error('clearTimeout has not been defined');
}
(function () {
    try {
        if (typeof setTimeout === 'function') {
            cachedSetTimeout = setTimeout;
        } else {
            cachedSetTimeout = defaultSetTimout;
        }
    } catch (e) {
        cachedSetTimeout = defaultSetTimout;
    }
    try {
        if (typeof clearTimeout === 'function') {
            cachedClearTimeout = clearTimeout;
        } else {
            cachedClearTimeout = defaultClearTimeout;
        }
    } catch (e) {
        cachedClearTimeout = defaultClearTimeout;
    }
} ())
function runTimeout(fun) {
    if (cachedSetTimeout === setTimeout) {
        //normal enviroments in sane situations
        return setTimeout(fun, 0);
    }
    // if setTimeout wasn't available but was latter defined
    if ((cachedSetTimeout === defaultSetTimout || !cachedSetTimeout) && setTimeout) {
        cachedSetTimeout = setTimeout;
        return setTimeout(fun, 0);
    }
    try {
        // when when somebody has screwed with setTimeout but no I.E. maddness
        return cachedSetTimeout(fun, 0);
    } catch(e){
        try {
            // When we are in I.E. but the script has been evaled so I.E. doesn't trust the global object when called normally
            return cachedSetTimeout.call(null, fun, 0);
        } catch(e){
            // same as above but when it's a version of I.E. that must have the global object for 'this', hopfully our context correct otherwise it will throw a global error
            return cachedSetTimeout.call(this, fun, 0);
        }
    }


}
function runClearTimeout(marker) {
    if (cachedClearTimeout === clearTimeout) {
        //normal enviroments in sane situations
        return clearTimeout(marker);
    }
    // if clearTimeout wasn't available but was latter defined
    if ((cachedClearTimeout === defaultClearTimeout || !cachedClearTimeout) && clearTimeout) {
        cachedClearTimeout = clearTimeout;
        return clearTimeout(marker);
    }
    try {
        // when when somebody has screwed with setTimeout but no I.E. maddness
        return cachedClearTimeout(marker);
    } catch (e){
        try {
            // When we are in I.E. but the script has been evaled so I.E. doesn't  trust the global object when called normally
            return cachedClearTimeout.call(null, marker);
        } catch (e){
            // same as above but when it's a version of I.E. that must have the global object for 'this', hopfully our context correct otherwise it will throw a global error.
            // Some versions of I.E. have different rules for clearTimeout vs setTimeout
            return cachedClearTimeout.call(this, marker);
        }
    }



}
var queue = [];
var draining = false;
var currentQueue;
var queueIndex = -1;

function cleanUpNextTick() {
    if (!draining || !currentQueue) {
        return;
    }
    draining = false;
    if (currentQueue.length) {
        queue = currentQueue.concat(queue);
    } else {
        queueIndex = -1;
    }
    if (queue.length) {
        drainQueue();
    }
}

function drainQueue() {
    if (draining) {
        return;
    }
    var timeout = runTimeout(cleanUpNextTick);
    draining = true;

    var len = queue.length;
    while(len) {
        currentQueue = queue;
        queue = [];
        while (++queueIndex < len) {
            if (currentQueue) {
                currentQueue[queueIndex].run();
            }
        }
        queueIndex = -1;
        len = queue.length;
    }
    currentQueue = null;
    draining = false;
    runClearTimeout(timeout);
}

process.nextTick = function (fun) {
    var args = new Array(arguments.length - 1);
    if (arguments.length > 1) {
        for (var i = 1; i < arguments.length; i++) {
            args[i - 1] = arguments[i];
        }
    }
    queue.push(new Item(fun, args));
    if (queue.length === 1 && !draining) {
        runTimeout(drainQueue);
    }
};

// v8 likes predictible objects
function Item(fun, array) {
    this.fun = fun;
    this.array = array;
}
Item.prototype.run = function () {
    this.fun.apply(null, this.array);
};
process.title = 'browser';
process.browser = true;
process.env = {};
process.argv = [];
process.version = ''; // empty string to avoid regexp issues
process.versions = {};

function noop() {}

process.on = noop;
process.addListener = noop;
process.once = noop;
process.off = noop;
process.removeListener = noop;
process.removeAllListeners = noop;
process.emit = noop;
process.prependListener = noop;
process.prependOnceListener = noop;

process.listeners = function (name) { return [] }

process.binding = function (name) {
    throw new Error('process.binding is not supported');
};

process.cwd = function () { return '/' };
process.chdir = function (dir) {
    throw new Error('process.chdir is not supported');
};
process.umask = function() { return 0; };

},{}]},{},[1])(1)
});
