import { useState } from 'react';
import Item from './items';
import { VersionData } from '../../tools/parser/types';
import DetailModal from '../detail/detail';

interface ListProps {
    data?: VersionData;
    titles: string[];
}

const List: React.FC<ListProps> = ({ data, titles }) => {
    const [modalOpen, setModalOpen] = useState(false);
    const [name, setDetailName] = useState<string>();
    return (
        <>
            <DetailModal
                modalOpen={modalOpen}
                setModalOpen={setModalOpen}
                data={data}
                name={name}
            />
            {data ? (
                <div className="grid grid-cols-1 gap-4 px-8 py-16 text-gray-600 md:px-32 lg:grid-cols-2">
                    {titles.map((key, index) => {
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
                    })}
                </div>
            ) : (
                <p>Loading...</p>
            )}
        </>
    );
};

export default List;
