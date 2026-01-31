from __future__ import annotations
from dataclasses import dataclass, field
from typing import Any, Dict, Generic, List, Optional, TypeVar, Protocol

T = TypeVar("T")

class Decodable(Protocol):
    def from_dict(self, d: Dict[str, Any]) -> Any: ...

@dataclass
class VdlError:
    code: str
    message: str
    details: Optional[Dict[str, Any]] = None

    def to_dict(self) -> Dict[str, Any]:
        result: Dict[str, Any] = {
            "code": self.code,
            "message": self.message,
        }
        if self.details is not None:
            result["details"] = self.details
        return result

    @staticmethod
    def from_dict(d: Dict[str, Any]) -> VdlError:
        return VdlError(
            code=d.get("code", ""),
            message=d.get("message", ""),
            details=d.get("details"),
        )

@dataclass
class Response(Generic[T]):
    result: Optional[T] = None
    error: Optional[VdlError] = None

    def to_dict(self) -> Dict[str, Any]:
        result: Dict[str, Any] = {}
        if self.result is not None:
            if hasattr(self.result, "to_dict"):
                result["result"] = self.result.to_dict()  # type: ignore
            else:
                result["result"] = self.result
        if self.error is not None:
            result["error"] = self.error.to_dict()
        return result

    @staticmethod
    def from_dict(d: Dict[str, Any], type_hook: Optional[Decodable] = None) -> Response[T]:
        result_val = d.get("result")
        if result_val is not None and type_hook is not None:
            if hasattr(type_hook, "from_dict"):
                result_val = type_hook.from_dict(result_val)
            else:
                # Basic types or list/dict
                pass 

        error_val = d.get("error")
        error = VdlError.from_dict(error_val) if error_val else None
        
        return Response(result=result_val, error=error)
