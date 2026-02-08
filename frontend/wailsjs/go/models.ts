export namespace diff {
	
	export class DiffNode {
	    path: string;
	    type: string;
	    left?: any;
	    right?: any;
	    children?: DiffNode[];
	
	    static createFrom(source: any = {}) {
	        return new DiffNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.type = source["type"];
	        this.left = source["left"];
	        this.right = source["right"];
	        this.children = this.convertValues(source["children"], DiffNode);
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
	export class DiffStats {
	    added: number;
	    removed: number;
	    changed: number;
	    equal: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.added = source["added"];
	        this.removed = source["removed"];
	        this.changed = source["changed"];
	        this.equal = source["equal"];
	    }
	}
	export class DiffResult {
	    root: DiffNode;
	    stats: DiffStats;
	
	    static createFrom(source: any = {}) {
	        return new DiffResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = this.convertValues(source["root"], DiffNode);
	        this.stats = this.convertValues(source["stats"], DiffStats);
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

export namespace loganalyzer {
	
	export class ValueFrequency {
	    value: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new ValueFrequency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.count = source["count"];
	    }
	}
	export class PathSummary {
	    path: string;
	    count: number;
	    objectHits: number;
	    distinctCount: number;
	    topValues: ValueFrequency[];
	
	    static createFrom(source: any = {}) {
	        return new PathSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.count = source["count"];
	        this.objectHits = source["objectHits"];
	        this.distinctCount = source["distinctCount"];
	        this.topValues = this.convertValues(source["topValues"], ValueFrequency);
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
	export class AnalysisResult {
	    paths: PathSummary[];
	    totalLines: number;
	    jsonLines: number;
	    skippedLines: number;
	    totalPaths: number;
	    totalPathOccurs: number;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.paths = this.convertValues(source["paths"], PathSummary);
	        this.totalLines = source["totalLines"];
	        this.jsonLines = source["jsonLines"];
	        this.skippedLines = source["skippedLines"];
	        this.totalPaths = source["totalPaths"];
	        this.totalPathOccurs = source["totalPathOccurs"];
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
	export class ComparisonStats {
	    totalPaths: number;
	    addedPaths: number;
	    removedPaths: number;
	    changedPaths: number;
	    equalPaths: number;
	    totalCountDelta: number;
	    totalObjectsDelta: number;
	    totalDistinctDelta: number;
	
	    static createFrom(source: any = {}) {
	        return new ComparisonStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalPaths = source["totalPaths"];
	        this.addedPaths = source["addedPaths"];
	        this.removedPaths = source["removedPaths"];
	        this.changedPaths = source["changedPaths"];
	        this.equalPaths = source["equalPaths"];
	        this.totalCountDelta = source["totalCountDelta"];
	        this.totalObjectsDelta = source["totalObjectsDelta"];
	        this.totalDistinctDelta = source["totalDistinctDelta"];
	    }
	}
	export class PathComparison {
	    path: string;
	    status: string;
	    left?: PathSummary;
	    right?: PathSummary;
	    countDelta: number;
	    objectsDelta: number;
	    distinctDelta: number;
	
	    static createFrom(source: any = {}) {
	        return new PathComparison(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.status = source["status"];
	        this.left = this.convertValues(source["left"], PathSummary);
	        this.right = this.convertValues(source["right"], PathSummary);
	        this.countDelta = source["countDelta"];
	        this.objectsDelta = source["objectsDelta"];
	        this.distinctDelta = source["distinctDelta"];
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
	export class ComparisonResult {
	    comparisons: PathComparison[];
	    stats: ComparisonStats;
	    leftFile: string;
	    rightFile: string;
	
	    static createFrom(source: any = {}) {
	        return new ComparisonResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.comparisons = this.convertValues(source["comparisons"], PathComparison);
	        this.stats = this.convertValues(source["stats"], ComparisonStats);
	        this.leftFile = source["leftFile"];
	        this.rightFile = source["rightFile"];
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

export namespace main {
	
	export class FileResult {
	    path: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new FileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.content = source["content"];
	    }
	}
	export class LogFileResult {
	    path: string;
	    result?: loganalyzer.AnalysisResult;
	
	    static createFrom(source: any = {}) {
	        return new LogFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.result = this.convertValues(source["result"], loganalyzer.AnalysisResult);
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
	export class NormalizeOptions {
	    sortKeys: boolean;
	    normalizeNumbers: boolean;
	    trimStrings: boolean;
	    nullEqualsAbsent: boolean;
	    sortArrays: boolean;
	    sortArraysByKey: string;
	
	    static createFrom(source: any = {}) {
	        return new NormalizeOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sortKeys = source["sortKeys"];
	        this.normalizeNumbers = source["normalizeNumbers"];
	        this.trimStrings = source["trimStrings"];
	        this.nullEqualsAbsent = source["nullEqualsAbsent"];
	        this.sortArrays = source["sortArrays"];
	        this.sortArraysByKey = source["sortArraysByKey"];
	    }
	}

}

export namespace paths {
	
	export class PathInfo {
	    path: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new PathInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.count = source["count"];
	    }
	}
	export class PathResult {
	    paths: PathInfo[];
	    totalPaths: number;
	    totalLeafs: number;
	
	    static createFrom(source: any = {}) {
	        return new PathResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.paths = this.convertValues(source["paths"], PathInfo);
	        this.totalPaths = source["totalPaths"];
	        this.totalLeafs = source["totalLeafs"];
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

