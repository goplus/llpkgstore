import { useMemo } from 'react';
import ReactPaginate from 'react-paginate';

interface PaginationProps {
    setItemOffset: (offset: number) => void;
    itemOffset: number;
    itemCount: number;
    pageSize: number;
}

const Pagination: React.FC<PaginationProps> = ({
    setItemOffset,
    itemOffset,
    itemCount,
    pageSize,
}) => {
    const handlePageClick = (data: { selected: number }) => {
        console.log(data.selected);
        setItemOffset(data.selected);
    };
    const pageCount = useMemo(() => {
        return itemCount ? Math.ceil(itemCount / pageSize) : 0;
    }, [pageSize, itemCount]);
    return (
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
    );
};

export default Pagination;
