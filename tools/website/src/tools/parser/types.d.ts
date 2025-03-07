export interface VersionData {
    [key: string]: {
        versions: {
            c: string;
            go: string[];
        }[];
    };
}
