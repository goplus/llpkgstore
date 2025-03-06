import { VersionData } from './types';

export const titleParser = (
    data?: VersionData,
    query?: string,
    page: number = 0,
    pageSize: number = 10,
): { data: string[]; totalCount: number } => {
    if (!data) return { data: [], totalCount: 0 };
    query = query?.trim();
    const parsedData = Object.keys(data).filter((key) => {
        return query ? key.includes(query) : true;
    });

    const startIndex = page * pageSize;
    const endIndex = startIndex + pageSize;

    return {
        data: parsedData.slice(startIndex, endIndex),
        totalCount: parsedData.length,
    };
};

export const versionParser = (
    data: VersionData[string],
    queryOrigin: string = '',
    queryMapped: string = '',
    page: number = 0,
    pageSize: number = 10,
): VersionData[string]['versions'] => {
    const filteredVersions = data.versions.filter((ver) => {
        let flag = false;
        if (queryMapped.trim())
            ver.converted.forEach((con) => {
                if (con.includes(queryMapped.trim())) flag = true;
            });
        else flag = true;
        return (
            (queryOrigin.trim()
                ? ver.original.includes(queryOrigin.trim())
                : true) && flag
        );
    });

    const startIndex = page * pageSize;
    const endIndex = startIndex + pageSize;

    return filteredVersions.slice(startIndex, endIndex);
};
