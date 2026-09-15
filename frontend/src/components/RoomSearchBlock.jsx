import { CellSimple, CellList, Container, Flex, SearchInput } from "@maxhub/max-ui";
import { LuDoorClosed } from "react-icons/lu";
import { useState } from 'react';
import '../styles/RoomSearch.css';

const RoomSearchBlock = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [searchResults, setSearchResults] = useState([]);
    const [isSearching, setIsSearching] = useState(false);

    const handleSearch = (value) => {
        setSearchTerm(value);
        if (value.length > 0) {
            setIsSearching(true);
            setTimeout(() => {
                setSearchResults(value
                    ? Array.from({ length: 5 }, (_, i) => ({
                        number: `${value}${i > 0 ? String.fromCharCode(96 + i) : ''}`,
                        floor: Math.floor(Math.random() * 5) + 1,
                        status: Math.random() > 0.5 ? 'normal' : 'warning'
                    }))
                    : []);
                setIsSearching(false);
            }, 300);
        } else {
            setSearchResults([]);
            setIsSearching(false);
        }
    };

    return (
        <div className="mt-2">
            <Container>
                <span className="mb-1 text-secondary">
                    Поиск производится автоматически при вводе
                </span>
                
                <SearchInput 
                    value={searchTerm}
                    onChange={(e) => handleSearch(e.target.value)}
                    placeholder="Введите номер комнаты"
                    className="mb-1"
                />

                <CellList mode="island" filled={false}>
                    {searchResults.length > 0 ? (
                        searchResults.map((room, i) => (
                            <CellSimple
                                key={i}
                                title={`Комната ${room.number}`}
                                subtitle={`Этаж ${room.floor}`}
                                before={<LuDoorClosed />}
                                showChevron
                                asChild
                            >
                                <a href={`/room/employ/${room.number}`} />
                            </CellSimple>
                        ))
                    ) : (
                        <div style={{ padding: '12px 0', color: 'var(--text-secondary, #888)', textAlign: 'center' }}>
                            {searchTerm ? (
                                isSearching ? 'Поиск...' : 'Ничего не найдено'
                            ) : (
                                'Введите номер комнаты для поиска'
                            )}
                        </div>
                    )}
                </CellList>
            </Container>
        </div>
    );
}

export default RoomSearchBlock;
