<script>
  import logo from './assets/images/logo-universal.png'
  import {Greet} from '../wailsjs/go/main/App.js'
  import {Login} from "../wailsjs/go/main/App.js";
  import {GetSlotList} from "../wailsjs/go/main/App.js";
  import {GetScanRedisKey} from "../wailsjs/go/main/App.js";
  import {GetRedisKeyData} from "../wailsjs/go/main/App.js";
  import Modal from './modal/Modal.svelte';
  import { fly } from 'svelte/transition';
  import {LogPrint} from "../wailsjs/runtime/runtime.js";

  class NodeInfo {

    constructor(strIpAndPort) {
      const strIpAndPortArr = strIpAndPort.split(":");
      this.nodeIpAndPort = strIpAndPort;
      this.ip = strIpAndPortArr[0];
      this.port = parseInt(strIpAndPortArr[1]);
      this.redisKeys = [];
      this.scanCursor = 0;
      this.scanFinish = false;
    }

    getRedisKeys(){
      return this.redisKeys;
    }

    getNodeIpAndPort(){
      return this.nodeIpAndPort;
    }

    getScanCursor(){
      return this.scanCursor;
    }

    isScanFinish(){
      return this.scanFinish;
    }
    setScanFinish(scanFinished){
      this.scanFinish = scanFinished;
    }

    addRedisKey(keys){
      this.redisKeys.push(...keys);
    }

    setScanCursor(pScanCursor){
        this.scanCursor = pScanCursor;
    }

  }


  let isLogin = false
  // let isLogin = true
  let resultText = "Please enter your name below 👇"
  let ipAndPort
  let pass
  let nodeIpInfoMap = {}
  let nodeIpList = []


  function getSlotList() {
    GetSlotList().then(ipAndPortList =>{

      for( let nodeIpPort of ipAndPortList){
        LogPrint(nodeIpPort)
        let nodeInfo = new NodeInfo(nodeIpPort);
        nodeIpInfoMap[nodeIpPort] = nodeInfo;
      }

      nodeIpList = ipAndPortList
    })
  }

  function login() {
    Login(ipAndPort,pass).then(result =>{
      isLogin = result.success
      if( !result.success){
        resultText = result.errorMessage
      }else{
        resultText = "Success Login"
        getSlotList()
        resultText = "Success Login2"
      }


    } )
  }

  function getScanRedisKey(nodeIpAndPort){
    const nodeInfo = nodeIpInfoMap[nodeIpAndPort]
    if (nodeInfo == null) {
      return;
    }
    if ( nodeInfo.isScanFinish() ){
      return;
    }

    const strNodeAndIp = nodeInfo.getNodeIpAndPort();
    const scanCursor = nodeInfo.getScanCursor();


    GetScanRedisKey(strNodeAndIp,scanCursor).then( result =>{
      // LogPrint(result)
      if ( !result.success) {
        return;
      }

      nodeInfo.setScanFinish(result.finish)
      nodeInfo.setScanCursor(result.cursor);
      nodeInfo.addRedisKey(result.keys);

      nodeIpInfoMap[nodeIpAndPort] = nodeInfo
    })
  }

  function getRedisKeyData(nodeIpAndPort,redisKey){
    // const nodeInfo = nodeIpInfoMap[nodeIpAndPort]
    // if (nodeInfo == null) {
    //   return;
    // }
    // if ( nodeInfo.isScanFinish() ){
    //   return;
    // }
    //
    // const strNodeAndIp = nodeInfo.getNodeIpAndPort();
    // const scanCursor = nodeInfo.getScanCursor();
    //
    //
    // GetRedisKeyData(strNodeAndIp,scanCursor).then( result =>{
    //   // LogPrint(result)
    //   if ( !result.success) {
    //     return;
    //   }
    //
    //   nodeInfo.setScanFinish(result.finish)
    //   nodeInfo.setScanCursor(result.cursor);
    //   nodeInfo.addRedisKey(result.keys);
    //
    //   nodeIpInfoMap[nodeIpAndPort] = nodeInfo
    // })
  }

</script>

<main>
  <div class="parent-div">

    <div class="fix-top-panel-div">

    </div>

    <div class="wrap-left-panel-div">
      <div class="left-panel-div">
        <!--{#if nodeIpMap != null}-->
        <ul style="color: black">
          {#each nodeIpList as nodeIp}
            <a>{nodeIp}</a>
            {#if !nodeIpInfoMap[nodeIp].isScanFinish()}
              <button on:click={getScanRedisKey(nodeIp)}>+</button>
            {/if}
            {#each nodeIpInfoMap[nodeIp].getRedisKeys() as redisKey}
              <li style="margin-left: 2px;font-size: 10px; " >
                <a href="#">{redisKey}</a>
<!--                <a href="#" on:click={}>{redisKey}</a>-->
              </li>
            {/each}
          {/each}
        </ul>
        <!--{/if}-->
      </div>
    </div>



    <div class="content-panel-div">

      <div class="textarea-div">
        <textarea class="textarea-input">

        </textarea>
      </div>

      <div class="result-div">

      </div>

    </div>

  </div>




<!--  <img alt="Wails logo" id="logo" src="{logo}">-->
  {#if !isLogin}
    <Modal>
      <h2 slot="header">
        <small style="color: black">LogIn</small>
      </h2>

        <div style="margin-bottom: 10px">
          <input class="login-input-box" placeholder="IP And Port" autocomplete="off" bind:value={ipAndPort} id="ipAndPort" type="text"/>
        </div>

      <div style="margin-bottom: 10px">
          <input class="login-input-box" placeholder="Password" autocomplete="off" bind:value={pass} id="pass" type="text"/>
        </div>

      <div>
        <button class="login-btm-box" on:click={login}>login</button>
      </div>

      {#if resultText}
        <div style="margin-top: 10px; color: black" >
          {resultText}
        </div>
      {/if}
    </Modal>
  {/if}





</main>

<style>

  #logo {
    display: block;
    width: 50%;
    height: 50%;
    margin: auto;
    padding: 10% 0 0;
    background-position: center;
    background-repeat: no-repeat;
    background-size: 100% 100%;
    background-origin: content-box;
  }

  .result {
    height: 20px;
    line-height: 20px;
    margin: 1.5rem auto;
  }

  .input-box .btn {
    width: 60px;
    height: 30px;
    line-height: 30px;
    border-radius: 3px;
    border: none;
    margin: 0 0 0 20px;
    padding: 0 8px;
    cursor: pointer;
  }

  .input-box .btn:hover {
    background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
    color: #333333;
  }

  .input-box .input {
    border: 1px solid;
    border-radius: 7px;
    outline: none;
    height: 30px;
    width: 100%;
    line-height: 30px;
    padding: 0 10px;
    background-color: rgba(240, 240, 240, 1);
    -webkit-font-smoothing: antialiased;
  }

  .login-input-box{
    border: 1px solid;
    border-radius: 7px;
    outline: none;
    height: 30px;
    width: 100%;
    line-height: 30px;
    background-color: rgba(240, 240, 240, 1);
    -webkit-font-smoothing: antialiased;
  }

  .login-btm-box{
    border: 1px solid;
    border-radius: 7px;
    outline: none;
    height: 50px;
    width: 100%;
    line-height: 30px;
    padding: 0 10px;
    background-color: #03c75a;
    -webkit-font-smoothing: antialiased;
  }

  .input-box .input:hover {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }

  .input-box .input:focus {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }

  .parent-div{
    height: 100%;
    width: 100%;
    position: fixed;
  }
  .fix-top-panel-div {
    background: #333333;
    height: 10%;
    width: 100%;
    float: top;
    /*position: fixed;*/


    /*border-bottom: 1px solid #03c75a;*/
    /*!*top: 0;*!*/
    /*!*padding: 2rem 1rem 0.6rem;*!*/
    /*border-left: 1px solid #aaa;*/

    /*overflow-y: auto;*/

  }

  .wrap-left-panel-div {
    background: #fff;
    /*height: 100%;*/
    height:100%;
    width: 20%;
    float: left;
    position: relative;
  }

  .left-panel-div {
    /*background: #fff;*/
    /*height: 100%;*/
    height:90%;
    width: 100%;
    /*position: absolute;*/
    /*float: left;*/
    overflow-y: auto;
    position: absolute;

  }

  .content-panel-div{
    /*background: #333333;*/
    height: 95%;
    width:80%;
    float: right;
  }

  .textarea-div{
    /*background: #fff;*/
    /*height: 100%;*/
    border: 1px solid red;
    height:50%;
    width: 100%;
    position: relative;
  }
  .textarea-input{
    height:100%;
    width: 100%;
  }
  .result-div{
    /*background: #fff;*/
    /*height: 100%;*/
    border: 1px solid #03c75a;
    height:50%;
    width: 100%;
  }



</style>
