import { useEffect, useMemo, useState } from 'react';
import Header from './layout/header';
import List from './components/list/list';
import Search from './components/search';
import { titleParser } from './tools/parser/parser';
import { VersionData } from './tools/parser/types';
import './App.css';
import ReactPaginate from 'react-paginate';

function App() {
    const [search, setSearch] = useState('');
    const [data, setData] = useState<VersionData>();
    const [itemOffset, setItemOffset] = useState(1);
    const pageSize = 4;
    const searchResult = useMemo(
        () => titleParser(data, search, itemOffset, pageSize),
        [data, search, itemOffset],
    );
    const pageCount = useMemo(() => {
        return searchResult.totalCount
            ? Math.ceil(searchResult.totalCount / pageSize)
            : 0;
    }, [searchResult]);
    useEffect(() => {
        getVersionData().then((data) => {
            setData(data);
        });
    }, []);
    useEffect(() => {
        setItemOffset(0);
    }, [search]);
    const handlePageClick = (data: { selected: number }) => {
        console.log(data.selected);
        setItemOffset(data.selected);
    };
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
            <ReactPaginate
                breakLabel="..."
                activeClassName="text-blue-500"
                nextLabel="next >"
                onPageChange={handlePageClick}
                forcePage={itemOffset}
                pageRangeDisplayed={5}
                pageCount={pageCount}
                previousLabel="< previous"
                renderOnZeroPageCount={null}
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
