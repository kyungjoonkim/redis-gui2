// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {bmodel} from '../models';
import {rmodel} from '../models';
import {main} from '../models';

export function GetRedisKeyData(arg1:string,arg2:string,arg3:number,arg4:number):Promise<bmodel.RedisGetResModel>;

export function GetScanRedisKey(arg1:string,arg2:number):Promise<rmodel.RedisScanResult>;

export function GetSlotList():Promise<Array<string>>;

export function Greet(arg1:string):Promise<string>;

export function Login(arg1:string,arg2:string):Promise<main.LoginResult>;
