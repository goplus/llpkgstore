import { useState } from 'react';
import Item from './items';
import { titleParser } from '../../tools/parser/parser';
import DetailModal from './detail';

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
    return (
        <>
            <DetailModal modalOpen={modalOpen} setModalOpen={setModalOpen} data={data} name={name} />
            <div className="grid grid-cols-2 gap-4 text-gray-600 px-32 py-16">
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
