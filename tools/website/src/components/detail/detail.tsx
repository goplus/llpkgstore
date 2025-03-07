import { useEffect, useMemo, useState } from 'react';
import { VersionData } from '../../tools/parser/types';
import { versionParser } from '../../tools/parser/parser';
import Modal from '../modal';
import Title from './title';
import VersionItem from './items';
import AscendingImg from '../../assets/sortAscending.svg';
import DescendingImg from '../../assets/sortDescending.svg';
import Pagination from '../pagination';

interface DetailModalProps {
    modalOpen: boolean;
    setModalOpen: (open: boolean) => void;
    name?: string;
    data?: VersionData;
}

const DetailModal: React.FC<DetailModalProps> = ({
    modalOpen,
    setModalOpen,
    name,
    data,
}) => {
    const [originVersion, setOriginVersion] = useState('');
    const [mappedVersion, setMappedVersion] = useState('');
    const [itemOffset, setItemOffset] = useState(0);
    const [desc, setDesc] = useState(false);
    const [version, setVersion] = useState('latest');
    const pageSize = 5;
    const searchResult = useMemo(
        () =>
            data &&
            name &&
            versionParser(
                data[name],
                originVersion,
                mappedVersion,
                itemOffset,
                pageSize,
                desc,
            ),
        [data, originVersion, mappedVersion, desc, name, itemOffset],
    );

    useEffect(() => {
        setVersion('latest');
        setDesc(false);
    }, [name]);
    useEffect(() => {
        setItemOffset(0);
    }, [originVersion, mappedVersion]);
    return (
        <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)}>
            <div className="flex h-full flex-col">
                <Title name={name} version={version} setVersion={setVersion} />
                <div className="relative h-full overflow-auto px-4 pb-3">
                    {searchResult ? (
                        <>
                            <div className="sticky top-1 mt-5 mb-1 flex flex-row justify-around text-base">
                                <div className="mb-2 flex items-center gap-2 rounded-lg border border-gray-300 bg-white/70 px-2 backdrop-blur-lg">
                                    <input
                                        onChange={(e) =>
                                            setOriginVersion(e.target.value)
                                        }
                                        className="border-r border-gray-300 px-1 py-2 focus-visible:outline-0"
                                        placeholder="Original version"
                                    />
                                    <input
                                        onChange={(e) =>
                                            setMappedVersion(e.target.value)
                                        }
                                        className="px-1 py-2 focus-visible:outline-0"
                                        placeholder="Mapped version"
                                    />
                                </div>
                            </div>
                            <div className="mb-3 text-right">
                                <button
                                    className="cursor-pointer text-sm text-gray-500"
                                    onClick={() => setDesc(!desc)}
                                >
                                    <img
                                        src={
                                            desc ? DescendingImg : AscendingImg
                                        }
                                        className="inline-block w-3 align-middle"
                                    />
                                    <span className="align-middle">
                                        {desc ? 'Descending' : 'Ascending'}
                                    </span>
                                </button>
                            </div>
                            <div className="flex flex-col gap-2">
                                {searchResult.data.map((ver, index) => {
                                    return (
                                        <VersionItem
                                            key={index}
                                            ver={ver}
                                            setVersion={setVersion}
                                        />
                                    );
                                })}
                            </div>
                            <Pagination
                                className="mt-3"
                                setItemOffset={setItemOffset}
                                itemOffset={itemOffset}
                                itemCount={searchResult.totalCount}
                                pageSize={pageSize}
                            />
                            {searchResult.totalCount === 0 && (
                                <p>No results.</p>
                            )}
                        </>
                    ) : (
                        <p>No data.</p>
                    )}
                </div>
            </div>
        </Modal>
    );
};

export default DetailModal;
