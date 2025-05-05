import {ElMessage} from "element-plus";

export function MsgError(err: any): void {
    ElMessage({
        showClose: true,
        message: err,
        type: 'error',
    })
}

export function MsgInfo(info: any): void {
    ElMessage({
        showClose: true,
        message: info,
    })
}

export function MsgWarn(warn: any): void {
    ElMessage({
        showClose: true,
        message: warn,
        type: 'warning',
    })
}

export function MsgSuccess(success: any): void {
    ElMessage({
        showClose: true,
        message: success,
        type: 'success',
    })
}


export interface MsgContainer<T> {
    data?: T;
    err?: string;
}

function handleError(err: string): void {
    ElMessage({
        showClose: true,
        message: err,
        type: 'error',
    })
}

export function ProcessResult<T>(input: MsgContainer<T>): T | void {
    if (input.err && input.err.length > 0) {
        handleError(input.err);
        return;
    }

    if ('data' in input) {
        return input.data;
    }

    throw new Error('Neither error nor data found in input');
}

export interface ErrorMsgContainer {
    err?: string;
}

export function ProcessError(input: ErrorMsgContainer): boolean {
    if (input.err && input.err.length > 0) {
        handleError(input.err);
        return false;
    }
    return true;
}

export function Sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export function NewLoadingContext(text?: string): LoadingContext {
    const lc = new LoadingContext();
    if (text) {
        lc.constText = text
    } else {
        lc.constText = "Loading..."
    }
    return lc
}

export class LoadingContext {
    isLoading: boolean;
    text: string;
    constText: string = "Loading...";
    currentTask: Promise<void> = Promise.resolve();

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.isLoading = source["isLoading"];
        this.text = source["text"];
        this.constText = source["constText"];
        this.currentTask = Promise.resolve();
    }

    async Loading(fn: () => Promise<void>, text ?: string): Promise<void> {
        const task = this.currentTask.then(async () => {
            if (text) {
                this.text = text;
            } else {
                this.text = this.constText;
            }
            this.isLoading = true;
            try {
                return await fn();
            } catch (e) {
                MsgError(e);
            } finally {
                this.isLoading = false;
            }
        });
        this.currentTask = task.catch(() => {
        });
        return task;
    }
}

export function BytesToString(bs: Uint8Array): string {
    const decoder = new TextDecoder("utf-8");
    return decoder.decode(bs);
}

export function StringToBytes(str: string): Uint8Array {
    const encoder = new TextEncoder();
    return encoder.encode(str);
}

