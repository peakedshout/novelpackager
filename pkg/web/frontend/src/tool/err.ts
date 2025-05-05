export class Error {
    err: string;

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.err = source["err"];
    }
}

export function NewError(err: string): Error {
    return new Error({
        err: err,
    });
}

export function IsError(input: any): input is Error {
    return input instanceof Error || (typeof input === 'object' && 'string' === typeof input.err);
}

export function GetError(input: any): string | undefined {
    return IsError(input) ? input.err : undefined;
}

export function HandleError(input: any, fn: (err: string) => void): boolean {
    if (IsError(input)) {
        fn(input.err)
        return true
    }
    return false
}
