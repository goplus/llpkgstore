import { useEffect, useMemo, useState } from 'react';
import Header from './layout/header';
import List from './components/list/list';
import Search from './components/search';
import { titleParser } from './tools/parser/parser';
import { VersionData } from './tools/parser/types';
import Pagination from './components/pagination';
import './App.css';

function App() {
    const [search, setSearch] = useState('');
    const [data, setData] = useState<VersionData>();
    const [itemOffset, setItemOffset] = useState(0);
    const pageSize = 4;
    const searchResult = useMemo(
        () => titleParser(data, search, itemOffset, pageSize),
        [data, search, itemOffset],
    );
    useEffect(() => {
        getVersionData().then((data) => {
            setData(data);
        });
    }, []);
    useEffect(() => {
        setItemOffset(0);
    }, [search]);
    return (
        <>
            <Header />
            <Search
                query={search}
                setSearch={setSearch}
                dataNumber={data ? Object.keys(data).length : 0}
                resultNumber={searchResult.totalCount}
            />
            <List data={data} titles={searchResult.data} />
            <Pagination
                itemCount={searchResult.totalCount}
                pageSize={pageSize}
                setItemOffset={setItemOffset}
                itemOffset={itemOffset}
            />
        </>
    );
}

async function getVersionData(): Promise<VersionData> {
    const response = await fetch('./llpkgstore.json', {
        method: 'GET',
        headers: {
            'Cache-Control': 'no-cache',
        },
    });
    const data = await response.json();
    return data;
}

export default App;
