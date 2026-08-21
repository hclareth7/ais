export namespace config {

	export class Config {
	    theme: string;
	    sshKeyPaths: string[];
	    ignoreDirs: string[];
	    lastOpenedPath: string;
	    fontSize: number;
	    sidebarWidth: number;
	    recentPaths: string[];
	    selectedModel: string;
	    provider: string;
	    vertexProject: string;
	    vertexRegion: string;
	    zoomLevel: number;
	    opacity: number;
	    readingWidth: number;
	    readerRadius: number;
	    backgroundMode: string;

	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.sshKeyPaths = source["sshKeyPaths"];
	        this.ignoreDirs = source["ignoreDirs"];
	        this.lastOpenedPath = source["lastOpenedPath"];
	        this.fontSize = source["fontSize"];
	        this.sidebarWidth = source["sidebarWidth"];
	        this.recentPaths = source["recentPaths"];
	        this.selectedModel = source["selectedModel"];
	        this.provider = source["provider"];
	        this.vertexProject = source["vertexProject"];
	        this.vertexRegion = source["vertexRegion"];
	        this.zoomLevel = source["zoomLevel"];
	        this.opacity = source["opacity"];
	        this.readingWidth = source["readingWidth"];
	        this.readerRadius = source["readerRadius"];
	        this.backgroundMode = source["backgroundMode"];
	    }
	}

}

export namespace main {

	export class UISettings {
	    zoomLevel: number;
	    opacity: number;
	    readingWidth: number;
	    readerRadius: number;
	    backgroundMode: string;

	    static createFrom(source: any = {}) {
	        return new UISettings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zoomLevel = source["zoomLevel"];
	        this.opacity = source["opacity"];
	        this.readingWidth = source["readingWidth"];
	        this.readerRadius = source["readerRadius"];
	        this.backgroundMode = source["backgroundMode"];
	    }
	}

}

export namespace highlights {

	export class Highlight {
	    id: string;
	    filePath: string;
	    anchorText: string;
	    prefixContext: string;
	    suffixContext: string;
	    color: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new Highlight(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filePath = source["filePath"];
	        this.anchorText = source["anchorText"];
	        this.prefixContext = source["prefixContext"];
	        this.suffixContext = source["suffixContext"];
	        this.color = source["color"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

export namespace search {

    export class SearchResult {
        filePath: string;
        lineNumber: number;
        matchOffset: number;
        context: string;

        static createFrom(source: any = {}) {
            return new SearchResult(source);
        }

        constructor(source: any = {}) {
            if ('string' === typeof source) source = JSON.parse(source);
            this.filePath = source["filePath"];
            this.lineNumber = source["lineNumber"];
            this.matchOffset = source["matchOffset"];
            this.context = source["context"];
        }
    }

}

export namespace types {
	
	export class FileNode {
	    name: string;
	    path: string;
	    isDir: boolean;
	    children?: FileNode[];
	
	    static createFrom(source: any = {}) {
	        return new FileNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.children = this.convertValues(source["children"], FileNode);
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

