export namespace bmodel {
	
	export class RedisGetResModel {
	    dataType: string;
	    redisKey: string;
	    next: number;
	    values: any;
	
	    static createFrom(source: any = {}) {
	        return new RedisGetResModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataType = source["dataType"];
	        this.redisKey = source["redisKey"];
	        this.next = source["next"];
	        this.values = source["values"];
	    }
	}

}

export namespace main {
	
	export class LoginResult {
	    success: boolean;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.errorMessage = source["errorMessage"];
	    }
	}

}

export namespace rmodel {
	
	export class RedisScanResult {
	    success: boolean;
	    errorMessage: string;
	    keys: string[];
	    cursor: number;
	    finish: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RedisScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.errorMessage = source["errorMessage"];
	        this.keys = source["keys"];
	        this.cursor = source["cursor"];
	        this.finish = source["finish"];
	    }
	}

}

