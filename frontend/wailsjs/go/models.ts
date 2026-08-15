export namespace config {
	
	export class AppSettings {
	    piCommand: string;
	    piSettingsPath: string;
	    piModelsPath: string;
	    piSwitchConfigPath: string;
	    darkMode: boolean;
	    lastDefaultProviderId?: string;
	    lastDefaultModelId?: string;
	    workingDir: string;
	    lastUpdateCheckAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.piCommand = source["piCommand"];
	        this.piSettingsPath = source["piSettingsPath"];
	        this.piModelsPath = source["piModelsPath"];
	        this.piSwitchConfigPath = source["piSwitchConfigPath"];
	        this.darkMode = source["darkMode"];
	        this.lastDefaultProviderId = source["lastDefaultProviderId"];
	        this.lastDefaultModelId = source["lastDefaultModelId"];
	        this.workingDir = source["workingDir"];
	        this.lastUpdateCheckAt = source["lastUpdateCheckAt"];
	    }
	}
	export class AppState {
	    version: string;
	    providers: provider.ConfigTransport[];
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
	        this.providers = this.convertValues(source["providers"], provider.ConfigTransport);
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
	
	export class ModelTransport {
	    id: string;
	    name: string;
	    api?: string;
	    baseUrl?: string;
	    reasoning: boolean;
	    thinkingLevelMap?: Record<string, any>;
	    input?: string[];
	    cost?: Record<string, any>;
	    contextWindow?: number;
	    maxTokens?: number;
	    samplingParams?: Record<string, any>;
	    headers?: Record<string, string>;
	    compat?: Record<string, any>;
	    compatExtraFieldsJson?: string;
	    extraFieldsJson?: string;
	    revision?: string;
	    replaceDocument?: boolean;
	    originalId?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelTransport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.api = source["api"];
	        this.baseUrl = source["baseUrl"];
	        this.reasoning = source["reasoning"];
	        this.thinkingLevelMap = source["thinkingLevelMap"];
	        this.input = source["input"];
	        this.cost = source["cost"];
	        this.contextWindow = source["contextWindow"];
	        this.maxTokens = source["maxTokens"];
	        this.samplingParams = source["samplingParams"];
	        this.headers = source["headers"];
	        this.compat = source["compat"];
	        this.compatExtraFieldsJson = source["compatExtraFieldsJson"];
	        this.extraFieldsJson = source["extraFieldsJson"];
	        this.revision = source["revision"];
	        this.replaceDocument = source["replaceDocument"];
	        this.originalId = source["originalId"];
	    }
	}
	export class ConfigTransport {
	    id: string;
	    name: string;
	    type: string;
	    baseUrl: string;
	    apiKeyEnv: string;
	    apiKeyLiteral: string;
	    api: string;
	    proxy: string;
	    headerMode: string;
	    headers: Record<string, string>;
	    customHeaders?: Record<string, string>;
	    models: ModelTransport[];
	    host: string;
	    selectedModelId: string;
	    extraFieldsJson?: string;
	    modelsRevision?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigTransport(source);
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
	        this.headerMode = source["headerMode"];
	        this.headers = source["headers"];
	        this.customHeaders = source["customHeaders"];
	        this.models = this.convertValues(source["models"], ModelTransport);
	        this.host = source["host"];
	        this.selectedModelId = source["selectedModelId"];
	        this.extraFieldsJson = source["extraFieldsJson"];
	        this.modelsRevision = source["modelsRevision"];
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
	export class ModelListTransport {
	    models: ModelTransport[];
	    revision: string;

	    static createFrom(source: any = {}) {
	        return new ModelListTransport(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.models = this.convertValues(source["models"], ModelTransport);
	        this.revision = source["revision"];
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

export namespace updater {
	
	export class CheckResult {
	    currentVersion: string;
	    latestVersion: string;
	    hasUpdate: boolean;
	    releaseUrl: string;
	    assetUrl: string;
	    releaseNotes: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.hasUpdate = source["hasUpdate"];
	        this.releaseUrl = source["releaseUrl"];
	        this.assetUrl = source["assetUrl"];
	        this.releaseNotes = source["releaseNotes"];
	    }
	}

}

export namespace wsl {

	export class Detection {
	    detected: boolean;
	    distros: string[];

	    static createFrom(source: any = {}) {
	        return new Detection(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.detected = source["detected"];
	        this.distros = source["distros"];
	    }
	}
	export class PiDetection {
	    distro: string;
	    home: string;
	    piAvailable: boolean;
	    piPath: string;
	    piHome: string;
	    piHomeExists: boolean;
	    settingsExists: boolean;
	    modelsExists: boolean;

	    static createFrom(source: any = {}) {
	        return new PiDetection(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.distro = source["distro"];
	        this.home = source["home"];
	        this.piAvailable = source["piAvailable"];
	        this.piPath = source["piPath"];
	        this.piHome = source["piHome"];
	        this.piHomeExists = source["piHomeExists"];
	        this.settingsExists = source["settingsExists"];
	        this.modelsExists = source["modelsExists"];
	    }
	}

}

