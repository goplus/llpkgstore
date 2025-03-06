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
        <div className="flex justify-center pb-6">
            <ReactPaginate
                breakLabel="..."
                activeLinkClassName="text-white bg-blue-400 hover:bg-blue-400 cursor-default border border-blue-300"
                className="flex gap-2"
                pageLinkClassName="cursor-pointer hover:border hover:border-gray-300 min-w-8 min-h-8 flex justify-center items-center rounded-lg"
                nextLinkClassName="cursor-pointer hover:border hover:border-gray-300 min-w-8 min-h-8 flex justify-center items-center rounded-r-lg"
                previousLinkClassName="cursor-pointer hover:border hover:border-gray-300 min-w-8 min-h-8 flex justify-center items-center rounded-l-lg"
                nextLabel=">"
                disabledLinkClassName="hidden"
                onPageChange={handlePageClick}
                forcePage={itemOffset}
                pageRangeDisplayed={5}
                pageCount={pageCount}
                previousLabel="<"
                renderOnZeroPageCount={null}
            />
        </div>
    );
};

export default Pagination;
