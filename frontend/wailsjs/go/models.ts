export namespace config {
	
	export class AppSettings {
	    piCommand: string;
	    piSettingsPath: string;
	    piModelsPath: string;
	    piSwitchConfigPath: string;
	    lastDefaultProviderId?: string;
	    lastDefaultModelId?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.piCommand = source["piCommand"];
	        this.piSettingsPath = source["piSettingsPath"];
	        this.piModelsPath = source["piModelsPath"];
	        this.piSwitchConfigPath = source["piSwitchConfigPath"];
	        this.lastDefaultProviderId = source["lastDefaultProviderId"];
	        this.lastDefaultModelId = source["lastDefaultModelId"];
	    }
	}
	export class AppState {
	    version: string;
	    providers: provider.Config[];
	    selectedProviderId: string;
	    defaultProviderId: string;
	    defaultModelId: string;
	    settings: AppSettings;
	    logs: string[];
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.providers = this.convertValues(source["providers"], provider.Config);
	        this.selectedProviderId = source["selectedProviderId"];
	        this.defaultProviderId = source["defaultProviderId"];
	        this.defaultModelId = source["defaultModelId"];
	        this.settings = this.convertValues(source["settings"], AppSettings);
	        this.logs = source["logs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace pi {
	
	export class LaunchPreview {
	    command: string;
	    checklist: string[];
	
	    static createFrom(source: any = {}) {
	        return new LaunchPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.checklist = source["checklist"];
	    }
	}

}

export namespace provider {
	
	export class ModelInfo {
	    id: string;
	    name: string;
	    reasoning: boolean;
	    contextWindow?: number;
	    maxTokens?: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.reasoning = source["reasoning"];
	        this.contextWindow = source["contextWindow"];
	        this.maxTokens = source["maxTokens"];
	    }
	}
	export class Config {
	    id: string;
	    name: string;
	    type: string;
	    baseUrl: string;
	    apiKeyEnv: string;
	    apiKeyLiteral: string;
	    api: string;
	    proxy: string;
	    headers: Record<string, string>;
	    models: ModelInfo[];
	    host: string;
	    selectedModelId: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKeyEnv = source["apiKeyEnv"];
	        this.apiKeyLiteral = source["apiKeyLiteral"];
	        this.api = source["api"];
	        this.proxy = source["proxy"];
	        this.headers = source["headers"];
	        this.models = this.convertValues(source["models"], ModelInfo);
	        this.host = source["host"];
	        this.selectedModelId = source["selectedModelId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConnectionTestResult {
	    ok: boolean;
	    title: string;
	    lines: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.title = source["title"];
	        this.lines = source["lines"];
	    }
	}

}

export namespace system {
	
	export class EnvCheckResult {
	    name: string;
	    found: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.found = source["found"];
	        this.message = source["message"];
	    }
	}

}

