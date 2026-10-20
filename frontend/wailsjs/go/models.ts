export namespace config {
	
	export class PackData {
	    repo: string;
	    version: string;
	    extracted_files: string[];
	
	    static createFrom(source: any = {}) {
	        return new PackData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repo = source["repo"];
	        this.version = source["version"];
	        this.extracted_files = source["extracted_files"];
	    }
	}
	export class AppConfig {
	    game_path: string;
	    app_lang: string;
	    installed_packs: Record<string, PackData>;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_path = source["game_path"];
	        this.app_lang = source["app_lang"];
	        this.installed_packs = this.convertValues(source["installed_packs"], PackData, true);
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

