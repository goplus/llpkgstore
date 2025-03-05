import { useState } from 'react';
import Modal from '../modal';
import Item from './items';
import { titleParser, versionParser } from '../../tools/parser/parser';

export interface VersionData {
    [key: string]: {
        versions: {
            original: string;
            converted: string[];
        }[];
    };
}

interface ListProps {
    data?: VersionData;
    search: string;
}

const List: React.FC<ListProps> = ({ data, search }) => {
    const [modalOpen, setModalOpen] = useState(false);
    const [name, setDetailName] = useState<string>();
    const [originVersion, setOriginVersion] = useState('');
    const [convertVersion, setConvertVersion] = useState('');
    return (
        <>
            <Modal isOpen={modalOpen} onClose={() => setModalOpen(false)} title={name}>
                {/* <Details /> */}
                {data && name && (
                    <>
                        <div className="flex flex-row justify-around text-base my-2">
                            <div className="px-2 py-1 border border-gray-300 rounded-3xl flex items-center gap-2">
                                <input
                                    onChange={(e) => setOriginVersion(e.target.value)}
                                    className="focus-visible:outline-0 px-1"
                                    placeholder="Original version"
                                />
                            </div>
                            <div className="px-2 py-1 border border-gray-300 rounded-3xl flex items-center gap-2">
                                <input
                                    onChange={(e) => setConvertVersion(e.target.value)}
                                    className="focus-visible:outline-0 px-1"
                                    placeholder="Converted version"
                                />
                            </div>
                        </div>
                        {versionParser(data[name], originVersion, convertVersion, 1, 10).map((ver, index) => {
                            return (
                                <div key={index} className="flex flex-row gap-4 items-center">
                                    <span className="font-bold text-lg min-w-16 text-left">{ver.original}</span>
                                    <span>{ver.converted.join(' / ')}</span>
                                </div>
                            );
                        })}
                    </>
                )}
            </Modal>
            <div className="flex flex-col gap-4 text-gray-600 px-32 py-16">
                {data ? (
                    titleParser(data, search, 1, 10).map((key, index) => {
                        return (
                            <Item
                                key={key}
                                name={key}
                                data={data[key]}
                                index={index}
                                setInfo={setDetailName}
                                setModalOpen={setModalOpen}
                            />
                        );
                    })
                ) : (
                    <p>Loading...</p>
                )}
            </div>
        </>
    );
};

export default List;
